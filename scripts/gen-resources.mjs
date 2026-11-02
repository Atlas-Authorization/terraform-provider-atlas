#!/usr/bin/env node
// Generates the table-driven resources from the Atlas OpenAPI document:
//   internal/provider/generic_specs_gen.go   the genSpec table
//   docs/resources/<name>.md, docs/data-sources/<name>.md
//   examples/resources/atlas_<name>/resource.tf
//   internal/provider/generic_examples_gen_test.go   example configs used by the tests
//
// Usage: node scripts/gen-resources.mjs [openapi.json | https://api.atlasauth.net/openapi.json]
// Request and response shapes come from the document; only the choices in
// scripts/resources-manifest.mjs (names, paths, overrides) are written by hand.
import fs from 'node:fs';
import { manifest } from './resources-manifest.mjs';

const src = process.argv[2] || 'https://api.atlasauth.net/openapi.json';
const oas = src.startsWith('http') ? await (await fetch(src)).json() : JSON.parse(fs.readFileSync(src, 'utf8'));
const warn = [];

const deref = (x, depth = 0) => {
  if (x && x.$ref && depth < 20) return deref(oas.components.schemas[x.$ref.split('/').pop()], depth + 1);
  return x;
};
const flat = (sch) => {
  sch = deref(sch);
  if (!sch) return { properties: {}, required: [] };
  const parts = sch.allOf ? sch.allOf.map(flat) : [];
  const props = Object.assign({}, ...parts.map((p) => p.properties), sch.properties || {});
  const required = [].concat(...parts.map((p) => p.required), sch.required || []);
  return { properties: props, required };
};
const jsonBody = (op) => op?.requestBody?.content?.['application/json']?.schema;
const jsonResp = (op) => {
  if (!op) return undefined;
  const code = Object.keys(op.responses || {}).find((k) => k[0] === '2');
  return code ? op.responses[code]?.content?.['application/json']?.schema : undefined;
};
const typeOf = (x) => {
  x = deref(x) || {};
  let t = x.type;
  if (Array.isArray(t)) t = t.find((v) => v !== 'null');
  return t;
};
const kindOf = (x) => {
  x = deref(x) || {};
  if (x.oneOf || x.anyOf) {
    const alts = (x.oneOf || x.anyOf).map(deref).filter((a) => typeOf(a) !== 'null');
    if (alts.length === 1) return kindOf(alts[0]);
    return 'json';
  }
  const t = typeOf(x);
  if (t === 'string') return 'string';
  if (t === 'integer') return 'int';
  if (t === 'number') return 'float';
  if (t === 'boolean') return 'bool';
  if (t === 'array') {
    const it = deref(x.items) || {};
    if (typeOf(it) === 'string' && !it.enum) return 'strlist';
    if (typeOf(it) === 'string') return 'strlist';
    return 'json';
  }
  if (Array.isArray(x.type) && x.type.filter((v) => v !== 'null').length > 1) return 'json';
  return 'json';
};
const enumOf = (x) => {
  x = deref(x) || {};
  if (x.enum) return x.enum;
  if (x.type === 'array' && x.items) return enumOf(x.items);
  const alts = x.oneOf || x.anyOf;
  if (alts) for (const a of alts) { const e = enumOf(a); if (e) return e; }
  return undefined;
};
const snake = (s) => s.replace(/([a-z0-9])([A-Z])/g, '$1_$2').replace(/[-.]/g, '_').toLowerCase();
const RESERVED = new Set(['count', 'for_each', 'lifecycle', 'depends_on', 'provider', 'providers', 'connection', 'provisioner', 'id']);
const SECRET = /(secret|password|passphrase|private_key|credential|api_key|token|headers|bind_pw)/i;
const VOLATILE = /(^updated_at$|^last_|^next_|_count$|^status$|^state$|^attempts?$|^users$|^rows$|^built_|^valid_until$|^error$|^renew|^usage)/;

const clean = (s) => (s || '').replace(/\s+/g, ' ').trim();
const esc = (s) => s.replace(/\\/g, '\\\\').replace(/"/g, '\\"');
const go = (s) => '"' + esc(s).replace(/\n/g, ' ') + '"';

function describe(name, sch, extra = '') {
  sch = deref(sch) || {};
  let d = clean(sch.description);
  if (!d) d = `The \`${name}\` field.`;
  const en = enumOf(sch);
  if (en && !d.includes(en[0])) d += ' One of: ' + en.map((e) => '`' + e + '`').join(', ') + '.';
  if (typeOf(sch) === 'integer' && /_at$/.test(name) && !/epoch|ms/.test(d)) d += ' (epoch ms)';
  if (kindOf(sch) === 'json') d += ' A JSON document (use `jsonencode`).';
  return (d + ' ' + extra).trim();
}

const specs = [];
for (const m of manifest) {
  const kind = m.kind || 'coll';
  const itemOas = m.item;
  const vars = [...itemOas.matchAll(/\{([^}]+)\}/g)].map((x) => x[1]);
  if (kind === 'dataonly') {
    const it = oas.paths[itemOas];
    if (!it?.get) { warn.push(`${m.t}: no GET on ${itemOas}`); continue; }
    const props = flat(jsonResp(it.get)).properties;
    const attrs = [];
    const used = new Set(['id']);
    for (const [api, sch] of Object.entries(props)) {
      if (['object', 'deleted', 'id'].includes(api) || (m.skip || []).includes(api)) continue;
      let name = (m.rename && m.rename[api]) || snake(api);
      if (RESERVED.has(name)) name += '_value';
      if (used.has(name)) continue;
      used.add(name);
      attrs.push({ name, api, kind: (m.kinds && m.kinds[api]) || kindOf(sch), comp: true, sens: /(secret|password|private_key)/.test(api), desc: describe(api, sch) });
    }
    specs.push({ t: m.t, kind, itemGo: itemOas.replace(/\{[^}]+\}$/, '{id}'), itemOas, idField: 'id', listGo: '', match: '', write: '', update: '', getItem: true, del: false, pick: '', ds: true, init: '', revoke: '', noRead: false, desc: clean(m.desc || it.get.summary), scope: it.get['x-required-scope'], attrs, m, updated: false });
    continue;
  }
  if (!oas.paths[itemOas] && kind !== 'single') { warn.push(`${m.t}: item path ${itemOas} not in the OpenAPI document`); continue; }
  if (!oas.paths[itemOas] && kind === 'single') { warn.push(`${m.t}: path ${itemOas} missing`); continue; }
  const item = oas.paths[itemOas];
  const idVar = kind === 'coll' ? vars[vars.length - 1] : null;
  const parentVars = kind === 'coll' ? vars.slice(0, -1) : vars;
  const paramName = (v) => (m.params && m.params[v]) || snake(v);
  const collOas = kind === 'coll' ? itemOas.replace(/\/\{[^}]+\}$/, '') : null;
  function collGo() { return itemGo.replace(/\/\{id\}$/, ''); }
  const itemGo = (() => {
    let t = itemOas;
    for (const v of parentVars) t = t.replace('{' + v + '}', '{' + paramName(v) + '}');
    if (idVar) t = t.replace('{' + idVar + '}', '{id}');
    return t;
  })();
  const coll = collOas ? oas.paths[collOas] : null;
  if (kind === 'coll' && !coll) { warn.push(`${m.t}: collection ${collOas} missing`); continue; }

  // operations
  let writeM, writeOp, updM, updOp;
  if (kind === 'coll') {
    writeM = m.write || 'post'; writeOp = coll[writeM];
    updM = m.update === undefined ? (item.patch ? 'patch' : item.put ? 'put' : '') : m.update;
    updOp = updM ? item[updM] : undefined;
  } else {
    writeM = m.write || (item.put ? 'put' : item.post ? 'post' : 'patch'); writeOp = item[writeM];
    updM = writeM; updOp = writeOp;
  }
  if (!writeOp) { warn.push(`${m.t}: no ${writeM} operation`); continue; }
  const getItem = !!item.get && !m.noGet;
  const del = !!item.delete && !m.noDelete;

  const body = flat(jsonBody(writeOp));
  const updBody = updOp && updOp !== writeOp ? flat(jsonBody(updOp)) : null;
  // response props: prefer GET item, then the write response
  const respProps = {};
  const respSchemas = [];
  const getResp = item.get && jsonResp(item.get);
  if (getResp) respSchemas.push(flat(getResp));
  respSchemas.push(flat(jsonResp(writeOp)));
  if (kind === 'coll' && !getItem && coll.get) {
    const lr = flat(jsonResp(coll.get));
    const di = deref(lr.properties?.data);
    if (di?.items) respSchemas.push(flat(di.items));
  }
  if (m.pick === 'data0') {
    const di = deref(flat(jsonResp(writeOp)).properties.data);
    respSchemas.length = 0;
    if (di?.items) respSchemas.push(flat(di.items));
    if (coll?.get) { const lr = deref(flat(jsonResp(coll.get)).properties.data); if (lr?.items) respSchemas.push(flat(lr.items)); }
  }
  for (const r of respSchemas) Object.assign(respProps, r.properties);
  const hasTypedResp = Object.keys(respProps).filter((k) => k !== 'object').length > 0;

  const used = new Set(['id']);
  const attrs = [];
  const nameFor = (api) => {
    let n = (m.rename && m.rename[api]) || snake(api);
    if (RESERVED.has(n)) n += '_value';
    return n;
  };
  const addParams = () => {
    for (const v of parentVars) {
      const n = paramName(v);
      used.add(n);
      attrs.push({ name: n, api: n, kind: 'string', req: true, param: true, desc: `The ${n.replace(/_/g, ' ')} this object belongs to. Changing it forces replacement.` });
    }
  };
  addParams();
  const paramApis = new Set(parentVars);
  const skip = new Set([...(m.skip || []), 'object', 'deleted', 'id', ...parentVars.map(paramName)]);
  if (kind === 'keyed') for (const v of parentVars) skip.add(v);

  // writable
  const secretsDeclared = new Set(m.secrets || []);
  for (const [api, sch] of Object.entries(body.properties)) {
    if (skip.has(api) || (kind === 'coll' && api === 'id')) continue;
    const name = nameFor(api);
    if (used.has(name)) { warn.push(`${m.t}: duplicate attribute ${name}`); continue; }
    used.add(name);
    const k = (m.kinds && m.kinds[api]) || kindOf(sch);
    const inResp = api in respProps;
    const secretLike = secretsDeclared.has(api) || (SECRET.test(api) && !inResp) || (SECRET.test(api) && api === 'credentials');
    const wo = secretLike || (m.writeOnly || []).includes(api) || (!inResp && hasTypedResp && !m.echo);
    const req = body.required.includes(api) || (m.required || []).includes(api);
    const immut = !!updBody && !m.patchAll && !(api in updBody.properties) && kind === 'coll';
    attrs.push({
      name, api, kind: k, req, sens: secretLike, writeOnly: wo, immut, enum: enumOf(sch),
      desc: describe(api, sch, secretLike ? 'Write-only: sent to Atlas, never returned, and not detected as drift.' : wo ? 'Not returned by the API, so it is not detected as drift.' : '') + (immut ? ' Changing it forces replacement.' : ''),
    });
  }
  // response-only (computed)
  const writable = new Set(Object.keys(body.properties));
  const extra = m.extra || [];
  for (const [api, sch] of Object.entries(respProps)) {
    if (writable.has(api) || skip.has(api)) continue;
    const name = nameFor(api);
    if (used.has(name)) continue;
    used.add(name);
    const k = (m.kinds && m.kinds[api]) || kindOf(sch);
    const sens = SECRET.test(api) && !/^has_/.test(api) && !/_at$/.test(api);
    attrs.push({ name, api, kind: k, comp: true, sens, volatile: VOLATILE.test(api), desc: describe(api, sch, sens ? 'Returned once, on create.' : '') });
  }
  for (const e of extra) {
    const name = nameFor(e.api);
    if (used.has(name)) continue;
    used.add(name);
    attrs.push({ name, api: e.api, kind: e.kind || 'string', comp: true, volatile: VOLATILE.test(e.api), sens: !!e.sens, desc: e.desc || `The \`${e.api}\` field.` });
  }
  if (m.raw) {
    used.add(m.raw);
    attrs.push({ name: m.raw, api: '__raw', kind: 'json', comp: true, volatile: true, desc: 'The full object the API returned, as a JSON document.' });
  }
  // an untyped response that echoes the request: nothing more to compute
  if (kind === 'coll' && !attrs.some((a) => a.comp) && !m.echo && !hasTypedResp) warn.push(`${m.t}: untyped response`);
  let listGo = '', match = '';
  if (!getItem) {
    if (kind === 'coll' && coll.get) { listGo = collGo(); match = (m.idField || (idVar && /^(key|email|name)$/.test(idVar) ? idVar : 'id')); const lr = flat(jsonResp(coll.get)); if (!lr.properties.data) warn.push(`${m.t}: list response has no data array`); }
    else if (m.list) { listGo = m.list.replace(/\{([^}]+)\}/g, (_, v) => '{' + paramName(v) + '}'); match = m.match; }
    else if (!m.noRead) warn.push(`${m.t}: no way to read the object (set noRead)`);
  }

  const scope = writeOp['x-required-scope'];
  const summary = clean(writeOp.summary);
  specs.push({
    t: m.t, kind, itemGo, itemOas, collOas, listGo, match, idField: m.idField || (kind === 'coll' && idVar && idVar !== 'id' && /^(key|email|name)$/.test(idVar) ? idVar : 'id'),
    write: writeM.toUpperCase(), update: kind === 'coll' ? updM.toUpperCase() : (m.update === '' ? '' : writeM.toUpperCase()),
    getItem, del, pick: m.pick || '', ds: !!m.ds, init: m.init || '', revoke: m.revoke || '', noRead: !!m.noRead,
    desc: clean(m.desc || summary), scope, scopeRead: (item.get || coll?.get)?.['x-required-scope'],
    attrs, m, parentVars, paramName,
    updated: !!updM,
    methods: { write: writeM, upd: updM, del, getItem },
  });
}

// data-only specs
for (const m of manifest.filter((x) => false)) void m;

// ── emit Go ────────────────────────────────────────────────────────────────
const b = (v) => (v ? 'true' : 'false');
let goSrc = `// Code generated by scripts/gen-resources.mjs from the Atlas OpenAPI document. DO NOT EDIT.

package provider

func genResourceSpecs() []genSpec {
	return []genSpec{
`;
for (const s of specs) {
  goSrc += `		{
			Type: ${go(s.t)}, Kind: ${go(s.kind)}, Item: ${go(s.itemGo)}, IDField: ${go(s.idField)},
			Write: ${go(s.write)}, Update: ${go(s.updated ? s.update : '')}, GetItem: ${b(s.getItem)}, Delete: ${b(s.del)},
			Pick: ${go(s.pick)}, List: ${go(s.listGo || '')}, Match: ${go(s.match || '')}, DS: ${b(s.ds)}, Init: ${go(s.init)}, Revoke: ${go(s.revoke)}, NoRead: ${b(s.noRead)},
			Desc: ${go(s.desc + (s.scope ? ' Needs the \`' + s.scope + '\` scope.' : ''))},
			Attrs: []genAttr{
`;
  for (const a of s.attrs) {
    goSrc += `				{Name: ${go(a.name)}, API: ${go(a.api)}, Kind: ${go(a.kind)}${a.req ? ', Req: true' : ''}${a.comp ? ', Comp: true' : ''}${a.sens ? ', Sens: true' : ''}${a.writeOnly ? ', WriteOnly: true' : ''}${a.immut ? ', Immut: true' : ''}${a.param ? ', Param: true' : ''}${a.volatile ? ', Volatile: true' : ''}, Desc: ${go(a.desc)}},
`;
  }
  goSrc += `			},
		},
`;
}
goSrc += `	}
}
`;
fs.writeFileSync('internal/provider/generic_specs_gen.go', goSrc);

// ── docs + examples ──────────────────────────────────────────────────────────
const sample = (a, s) => {
  if (a.enum && a.kind === 'string') return '"' + a.enum[0] + '"';
  if (a.kind === 'bool') return 'true';
  if (a.kind === 'int') return '1';
  if (a.kind === 'float') return '1';
  if (a.kind === 'strlist') return '["example"]';
  if (a.kind === 'json') return 'jsonencode({})';
  if (/url|uri|endpoint|issuer|host/.test(a.api)) return '"https://example.com/' + a.api + '"';
  if (/email/.test(a.api)) return '"user@example.com"';
  return '"' + (a.sample || ('tf-' + s.t.replace(/_/g, '-'))) + '"';
};
fs.mkdirSync('docs/resources', { recursive: true });
fs.mkdirSync('docs/data-sources', { recursive: true });
const examples = {};
for (const s of specs) {
  if (s.kind === 'dataonly') {
    fs.writeFileSync(`docs/data-sources/${s.t}.md`, `# atlas_${s.t} (data source)

${s.desc}

Every attribute except \`id\` is computed.${s.scope ? ' Scope: \`' + s.scope + '\`.' : ''}

## Example Usage

\`\`\`hcl
data "atlas_${s.t}" "example" {
  id = "<id>"
}
\`\`\`

## Schema

### Required

- \`id\` (String) The id to look up.

### Read-Only

${s.attrs.map((a) => `- \`${a.name}\` (${{ string: 'String', bool: 'Boolean', int: 'Number', float: 'Number', strlist: 'List of String', json: 'String, JSON' }[a.kind]}${a.sens ? ', Sensitive' : ''}) ${a.desc}`).join('\n')}
`);
    fs.mkdirSync(`examples/data-sources/atlas_${s.t}`, { recursive: true });
    fs.writeFileSync(`examples/data-sources/atlas_${s.t}/data-source.tf`, `data "atlas_${s.t}" "example" {\n  id = "<id>"\n}\n`);
    continue;
  }
  const man = s.m;
  const req = s.attrs.filter((a) => a.req);
  const opt = s.attrs.filter((a) => !a.req && !a.comp);
  const ro = s.attrs.filter((a) => a.comp);
  const ex = man.example || req.map((a) => `  ${a.name} = ${(man.sample && man.sample[a.name]) || sample(a, s)}`).join('\n');
  const hcl = `resource "atlas_${s.t}" "example" {\n${ex}\n}\n`;
  examples[s.t] = hcl;
  fs.mkdirSync(`examples/resources/atlas_${s.t}`, { recursive: true });
  fs.writeFileSync(`examples/resources/atlas_${s.t}/resource.tf`, `# ${s.desc.split('. ')[0].replace(/\.$/, '')}.\n` + hcl);
  const line = (a) => `- \`${a.name}\` (${{ string: 'String', bool: 'Boolean', int: 'Number', float: 'Number', strlist: 'List of String', json: 'String, JSON' }[a.kind]}${a.sens ? ', Sensitive' : ''}) ${a.desc}`;
  const importDoc = (() => {
    const ps = s.attrs.filter((a) => a.param).map((a) => `<${a.name}>`);
    if (s.kind === 'single') return 'Import with any id; the object is read from the singleton path.\n\n```sh\nterraform import atlas_' + s.t + '.example ' + s.t + '\n```';
    if (s.kind === 'keyed') return 'The import id is `' + ps.join('/') + '`.\n\n```sh\nterraform import atlas_' + s.t + '.example ' + ps.join('/') + '\n```';
    return 'The import id is `' + [...ps, '<id>'].join('/') + '`.\n\n```sh\nterraform import atlas_' + s.t + '.example ' + [...ps, '<id>'].join('/') + '\n```';
  })();
  const routes = [];
  if (s.kind === 'coll') {
    routes.push(`\`${s.write} ${s.collOas}\` (create)`);
    if (s.getItem) routes.push(`\`GET ${s.itemOas}\` (read)`); else routes.push(`\`GET ${s.collOas}\` (read: the list is scanned for the id)`);
    if (s.updated) routes.push(`\`${s.update} ${s.itemOas}\` (update)`); else routes.push('no update route: every change replaces the object');
    if (s.revoke) routes.push(`\`POST ${s.itemOas}${s.revoke}\` (destroy)`); else if (s.del) routes.push(`\`DELETE ${s.itemOas}\` (destroy)`); else routes.push('no delete route: destroy forgets the object');
  } else {
    routes.push(`\`${s.write} ${s.itemOas}\` (create and update)`);
    routes.push(s.getItem ? `\`GET ${s.itemOas}\` (read)` : 'no read route: refresh keeps the last applied values');
    routes.push(s.del ? `\`DELETE ${s.itemOas}\` (destroy)` : 'no delete route: destroy forgets the object and leaves the setting in place');
  }
  const md = `# atlas_${s.t}

${s.desc}${s.scope ? '\n\nScopes: `' + (s.scopeRead || s.scope) + '` to read, `' + s.scope + '` to write.' : ''}

Backend API: ${routes.join('; ')}.

## Example Usage

\`\`\`hcl
${hcl}\`\`\`

## Schema

${req.length ? '### Required\n\n' + req.map(line).join('\n') + '\n\n' : ''}${opt.length ? '### Optional\n\n' + opt.map(line).join('\n') + '\n\n' : ''}### Read-Only

- \`id\` (String) Atlas object id.
${ro.map(line).join('\n')}

## Import

${importDoc}
${s.attrs.some((a) => a.writeOnly) ? '\nWrite-only attributes are not returned by the API, so they are empty after an import and must be set in configuration.\n' : ''}`;
  fs.writeFileSync(`docs/resources/${s.t}.md`, md.replace(/\n\n\n+/g, '\n\n'));
  if (s.ds) {
    const keys = s.attrs.filter((a) => a.param || false);
    fs.mkdirSync(`examples/data-sources/atlas_${s.t}`, { recursive: true });
    fs.writeFileSync(`docs/data-sources/${s.t}.md`, `# atlas_${s.t} (data source)

Reads an existing ${s.t.replace(/_/g, ' ')} by id. Every attribute except the lookup keys is computed; write-only secrets are not exposed.

## Example Usage

\`\`\`hcl
data "atlas_${s.t}" "example" {
${[...(s.kind === 'coll' ? ['  id = "<id>"'] : []), ...keys.map((a) => `  ${a.name} = "<${a.name}>"`)].join('\n')}
}
\`\`\`
`);
  }
}
fs.writeFileSync('internal/provider/generic_examples_gen_test.go', `// Code generated by scripts/gen-resources.mjs. DO NOT EDIT.

package provider

// genExampleConfigs maps a generated resource type to a minimal HCL example.
var genExampleConfigs = map[string]string{
${specs.filter((s) => s.kind !== 'dataonly').map((s) => `	${go(s.t)}: ${'`'}${examples[s.t].replace(/`/g, '')}${'`'},`).join('\n')}
}
`);
// docs/index.md: the generated tables live between markers
{
  const first = (d) => d.split(/(?<=\.)\s/)[0].replace(/\|/g, '/');
  const rows = specs.filter((s) => s.kind !== 'dataonly').map((s) => `| [\`atlas_${s.t}\`](resources/${s.t}.md) | ${first(s.desc)} |`);
  const drows = specs.filter((s) => s.ds || s.kind === 'dataonly').map((s) => `| [\`atlas_${s.t}\`](data-sources/${s.t}.md) | Look up ${s.t.replace(/_/g, ' ')} by id. |`);
  const block = `<!-- generated:start -->\n\n### More resources\n\n| Resource | Description |\n|----------|-------------|\n${rows.join('\n')}\n\n### More data sources\n\n| Data source | Description |\n|-------------|-------------|\n${drows.join('\n')}\n\n<!-- generated:end -->`;
  let idx = fs.readFileSync('docs/index.md', 'utf8');
  if (idx.includes('<!-- generated:start -->')) idx = idx.replace(/<!-- generated:start -->[\s\S]*<!-- generated:end -->/, block);
  else idx = idx.replace(/\n+$/, '\n') + '\n' + block + '\n';
  fs.writeFileSync('docs/index.md', idx);
}
console.log(`generated ${specs.length} resources`);
for (const w of warn) console.log('WARN', w);
