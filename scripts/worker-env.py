from pathlib import Path
allowed={'WORKER_API_BASE_URL','WORKER_TOKEN','WORKER_CONCURRENCY','WORKER_POLL_SECONDS','OPENCLAW_BASE_URL','OPENCLAW_GATEWAY_TOKEN','OPENCLAW_AGENT_ID','OPENCLAW_TIMEOUT_SECONDS'}
values={}
for line in Path('.env').read_text().splitlines():
 if '=' in line:
  k,v=line.split('=',1)
  if k in allowed:values[k]=v
p=Path('.local/worker.env');p.write_text('\n'.join(k+'='+v for k,v in values.items())+'\n');p.chmod(0o600)
print('Worker-only environment prepared; no database credential included.')
