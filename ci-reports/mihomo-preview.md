# Mihomo preview build report

Source commit: a6008bb6c0bc3f1ab9a9d2c811c2e252aaeb63ad

Run: https://github.com/alekoren/xkeen-panel/actions/runs/44325147092

## go: success

## node: success

## npm: success

~~~~text
npm warn deprecated node-domexception@1.0.0: Use your platform's native DOMException instead

added 409 packages, and audited 410 packages in 11s

124 packages are looking for funding
  run `npm fund` for details

28 vulnerabilities (2 low, 3 moderate, 22 high, 1 critical)

To address issues that do not require attention, run:
  npm audit fix

To address all issues (including breaking changes), run:
  npm audit fix --force

Run `npm audit` for details.

~~~~

## frontend: success

~~~~text

> xkeen-panel@0.0.0 build
> tsc -b && vite build

[36mvite v7.3.1 [32mbuilding client environment for production...[36m[39m
transforming...
[32m✓[39m 6121 modules transformed.
rendering chunks...
computing gzip size...
[2mdist/[22m[32mindex.html                 [39m[1m[2m  0.91 kB[22m[1m[22m[2m │ gzip:   0.41 kB[22m
[2mdist/[22m[2massets/[22m[35mindex-DjLrqkNL.css  [39m[1m[2m 48.42 kB[22m[1m[22m[2m │ gzip:   9.06 kB[22m
[2mdist/[22m[2massets/[22m[36mindex-DMLFBpDs.js   [39m[1m[2m420.23 kB[22m[1m[22m[2m │ gzip: 130.05 kB[22m
[32m✓ built in 3.76s[39m

~~~~

## tests: success

~~~~text
?   	xkeen-panel	[no test files]
ok  	xkeen-panel/internal/api	0.500s
ok  	xkeen-panel/internal/auth	0.338s
ok  	xkeen-panel/internal/coreupdate	(cached)
ok  	xkeen-panel/internal/geoip	(cached)
ok  	xkeen-panel/internal/keenetic	(cached)
ok  	xkeen-panel/internal/localtz	(cached)
ok  	xkeen-panel/internal/mihomo	(cached)
?   	xkeen-panel/internal/models	[no test files]
ok  	xkeen-panel/internal/monitor	(cached)
ok  	xkeen-panel/internal/server	0.004s
ok  	xkeen-panel/internal/sse	(cached)
ok  	xkeen-panel/internal/updater	(cached)
?   	xkeen-panel/internal/version	[no test files]
ok  	xkeen-panel/internal/xkeen	0.719s

~~~~

## binary: success

~~~~text

~~~~

