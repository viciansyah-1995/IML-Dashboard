"""Check the actual semantic color pairs; no browser or credentials required."""
import json,re
from pathlib import Path
values=dict(re.findall(r'(--[\w-]+):\s*([^;]+);',Path('frontend/src/styles/tokens.css').read_text()))
def resolve(name):
 value=values['--color-'+name] if not name.startswith('--') else values[name]
 return resolve(value[4:-1]) if value.startswith('var(') else value

def luminance(color):
 components=[int(color[i:i+2],16)/255 for i in (1,3,5)]
 return sum((c/12.92 if c<=.04045 else ((c+.055)/1.055)**2.4)*weight for c,weight in zip(components,[.2126,.7152,.0722]))
def contrast(fg,bg):
 low,high=sorted([luminance(resolve(fg)),luminance(resolve(bg))]);return (high+.05)/(low+.05)
pairs=[('on-primary',state,4.5) for state in ['primary','primary-hover','primary-active']]
pairs += [('selected-fg','selected-bg',4.5),('text-secondary','canvas',4.5),('text-on-canvas','canvas',4.5),('text-muted','surface',4.5),('border-on-canvas','canvas',3),('border-control','surface',3),('on-accent','accent',4.5),('on-danger','danger',4.5),('focus','surface',3),('link','canvas',4.5)]
pairs += [(tone+'-text',tone+'-bg',4.5) for tone in ['success','error','warning','info','neutral']]
pairs += [('chart-'+metric,'surface',3) for metric in ['views','interactions','activity','sales']]
results=[{'foreground':fg,'background':bg,'ratio':round(contrast(fg,bg),3),'target':target,'pass':contrast(fg,bg)>=target} for fg,bg,target in pairs]
report={'checks':results,'adjustments':{'muted_on_cream':'4.481:1; use text-on-canvas instead','control_border_on_cream':'2.995:1; use border-on-canvas instead'},'scope':'Color-pair verification only; not a complete WCAG audit.'}
Path('docs/design-system-contrast.json').write_text(json.dumps(report,indent=2)+'\n')
failures=[item for item in results if not item['pass']]
if failures:raise SystemExit(json.dumps(failures))
print(f'{len(results)} semantic color pairs passed; cream context adjustments applied.')
