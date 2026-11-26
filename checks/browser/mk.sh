#!/bin/sh
# refresh the test copy of the page and re-add the test hook
cp /home/claude/maesagetabbl/web/index.html /home/claude/test/index.html
python3 -I - <<'PY'
p='/home/claude/test/index.html'; s=open(p).read()
hook="window.__maesage = { S, T, start, begin, lobby, get G() { return G; }, P, toS, Z, SETS, HOLES, stop, enterTable, enterWatch, motor, get mot() { return mot; }, get feeling() { return feeling; } };\n"
i=s.rindex('})();\n</script>'); open(p,'w').write(s[:i]+hook+s[i:])
PY
