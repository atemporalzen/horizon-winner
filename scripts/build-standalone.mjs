import {readFileSync,writeFileSync,mkdirSync} from "node:fs";
import {fileURLToPath} from "node:url";
const root=new URL("../",import.meta.url);
const template=readFileSync(new URL("web/amaze.template.html",root),"utf8");
const engine=readFileSync(new URL("web/engine.js",root),"utf8").replaceAll("export ","");
const style=readFileSync(new URL("web/style.css",root),"utf8");
const result=template.replace("/*__STYLE__*/",()=>style).replace("/*__ENGINE__*/",()=>engine);
mkdirSync(new URL("html/",root),{recursive:true});
for(const name of ["amaze.html"]){
const file=new URL(name,root);
if(process.argv.includes("--check")){
  if(readFileSync(file,"utf8")!==result)throw new Error("amaze.html is stale; run npm run build:standalone");
}else{writeFileSync(file,result);console.log(`Built ${fileURLToPath(file)}`);}
}
const script=result.match(/<script>([\s\S]*?)<\/script>/)[1];
new Function(script); // Parse the embedded plain-HTTP script without executing it.
