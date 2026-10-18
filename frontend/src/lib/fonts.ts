// Add genuine webfont files to src/assets/fonts; Vite includes only files present.
const sources=import.meta.glob<string>('../assets/fonts/*.{woff2,woff}',{eager:true,query:'?url',import:'default'});
export function loadProjectFonts(){
 const entries=Object.entries(sources).filter(([path])=>/diatype/i.test(path)&&!/mono|condensed|compressed|expanded|extended|rounded/i.test(path));
 for(const [path,url] of entries){
  const name=path.split('/').pop()||'',style=/italic/i.test(name)?'italic':'normal';
  const weight=/variable|\bvf\b/i.test(name)?'100 900':/bold/i.test(name)?'700':/medium/i.test(name)?'500':/regular|italic/i.test(name)?'400':'';
  if(!weight)continue;
  const face=new FontFace('Diatype',`url("${url}")`,{weight,style,display:'swap'});
  document.fonts.add(face);
  void face.load().catch(()=>{document.fonts.delete(face)});
 }
}
