import {experiment,boundedUTF8} from "/engine.js";
const $=(id)=>document.getElementById(id);
let config,popup,abort,report,active=false,lastURL,objectURL;
const sleep=(ms,signal)=>new Promise((resolve,reject)=>{
  if(signal?.aborted){reject(new DOMException("Aborted","AbortError"));return;}
  const done=()=>{signal?.removeEventListener("abort",cancel);resolve();};
  const timer=setTimeout(done,ms);
  const cancel=()=>{clearTimeout(timer);signal?.removeEventListener("abort",cancel);reject(new DOMException("Aborted","AbortError"));};
  signal?.addEventListener("abort",cancel,{once:true});
});
function event(e){const li=document.createElement("li");li.textContent=`${(e.at_ms/1000).toFixed(1)}s · ${e.kind} · ${e.detail}`;$("events").appendChild(li);$("status").textContent=e.detail;$("state").textContent=e.kind.replaceAll("_"," ");}
async function json(url,options={}) {const r=await fetch(url,{cache:"no-store",...options});if(!r.ok)throw new Error(`Server returned ${r.status}`);return r.json();}
function isLocal(ip){return /^(127\.|10\.|192\.168\.|169\.254\.|172\.(1[6-9]|2\d|3[01])\.)/.test(ip)||ip==="::1"||/^(fc|fd|fe8|fe9|fea|feb)/i.test(ip);}
// randomUUID requires a secure context; rebinding entry pages are plain HTTP.
function nonce(){return Array.from(crypto.getRandomValues(new Uint8Array(16)),b=>b.toString(16).padStart(2,"0")).join("");}
async function prepare(){
  try{config=await json("/api/run");const t=config.session.target;$("target").textContent=`${t.label} · ${t.ip}:${t.port}${t.path}`;
    const entry=config.entry_ip;
    $("boundary").textContent=isLocal(entry)?"Local entry address: this run validates plumbing, not a public-to-local LNA boundary.":"Public entry address: record browser version and permissions alongside the result. A fetch error does not identify LNA on its own.";
    $("state").textContent="Ready";$("status").textContent="Start opens one visible test popup.";$("start").disabled=false;
  }catch(err){$("state").textContent="Session unavailable";$("status").textContent=err.message;}
}
async function probe(remainingMS){
  const controller=new AbortController();const propagate=()=>controller.abort();abort.signal.addEventListener("abort",propagate,{once:true});
  const timer=setTimeout(()=>controller.abort(new DOMException("Fetch comparison timed out","TimeoutError")),Math.min(config.request_timeout_ms,remainingMS));
  try{
    const url=new URL(config.session.target.path,location.origin);url.searchParams.set("__horizon_probe",nonce());
    const r=await fetch(url,{cache:"no-store",credentials:"omit",redirect:"error",signal:controller.signal});
    if(r.headers.get("X-Horizon-Origin")==="entry")return {kind:"entry"};
    // Bound allocations when comparing a configured target response.
    const reader=r.body?.getReader();if(!reader)return {kind:"other"};
    const decoder=new TextDecoder();let body="",size=0;
    try{while(size<config.max_response_bytes){const {done,value}=await reader.read();if(done)break;const chunk=value.subarray(0,config.max_response_bytes-size);size+=chunk.length;body+=decoder.decode(chunk,{stream:true});}body+=decoder.decode();}finally{await reader.cancel();}
    return {kind:body.includes(config.session.target.proof_marker)?"target":"other"};
  }finally{clearTimeout(timer);abort.signal.removeEventListener("abort",propagate);}
}
function navigate(){const url=new URL(config.session.target.path,location.origin);url.searchParams.set("__horizon_nav",nonce());lastURL=url;popup.location.href=url.href;}
function read(){
  const url=new URL(popup.location.href);
  if(url.origin!==location.origin)return {kind:"unreadable"};
  if(!lastURL||url.pathname!==lastURL.pathname||url.searchParams.get("__horizon_nav")!==lastURL.searchParams.get("__horizon_nav"))return {kind:"stale"};
  const doc=popup.document;if(doc.readyState!=="complete"||doc.querySelector("[data-horizon-entry]"))return {kind:"entry"};
  const text=doc.body?.textContent||"";
  if(text.includes("horizon-host-denied"))return {kind:"host_rejected"};
  if(!text.includes(config.session.target.proof_marker))return {kind:"other"};
  const html=doc.documentElement.outerHTML; // Serialize only after proof is found.
  const document=boundedUTF8(html,config.max_response_bytes);
  return {kind:"target",url:url.href,document,truncated:document.length<html.length};
}
$("start").addEventListener("click",async()=>{
  if(active||!config)return;
  // Opening happens synchronously within the click, before any await.
  popup=window.open("/popup",`horizon-${config.session.id}`,"popup,width=650,height=480");
  if(!popup){$("state").textContent="Popup blocked";$("status").textContent="Allow this visible popup and click Start again.";return;}
  active=true;$("start").disabled=true;$("stop").disabled=false;$("events").replaceChildren();abort=new AbortController();
  report={schema_version:1,session_id:config.session.id,target:config.session.target,started_at:new Date().toISOString(),browser:navigator.userAgent,origin:location.origin,entry_ip:config.entry_ip,research_url:"https://blog.criticalthinkingpodcast.io/p/hackernotes-ep-193-browser-quirks-galore-with-j0r1an",validation:"browser observation; LNA cause requires console/network evidence"};
  const started=performance.now();const clock=setInterval(()=>$("elapsed").textContent=`Elapsed ${Math.floor((performance.now()-started)/1000)}s / ${config.run_timeout_seconds}s`,1000);
  try{
    // Arm before DNS changes; repeats are idempotent and cannot postpone a run.
    await json("/api/arm",{method:"POST",headers:{"X-Horizon-Run":config.run_key},signal:abort.signal});
    const result=await experiment({timeoutMS:config.run_timeout_seconds*1000,holdMS:config.rebind_after_seconds*1000,pollMS:config.poll_interval_ms,navigationRetryMS:Math.max(5000,config.request_timeout_ms+1000)},{now:()=>performance.now(),signal:abort.signal,sleep,closed:()=>!popup||popup.closed,probe,navigate,read,event});
    report={...report,...result};$("state").textContent=result.outcome.replaceAll("_"," ");if(result.document)$("response").textContent=result.document;
  }catch(err){report={...report,outcome:abort.signal.aborted?"stopped":"setup_error",error:String(err.message||err)};$("state").textContent=report.outcome.replaceAll("_"," ");$("status").textContent=report.error;}
  finally{clearInterval(clock);active=false;$("stop").disabled=true;$("download").disabled=false;config.run_key="";}
});
$("stop").addEventListener("click",()=>{abort?.abort();popup?.close();});
$("download").addEventListener("click",()=>{if(!report)return;if(objectURL)URL.revokeObjectURL(objectURL);objectURL=URL.createObjectURL(new Blob([JSON.stringify(report,null,2)],{type:"application/json"}));const a=document.createElement("a");a.href=objectURL;a.download=`horizon-${config.session.id}.json`;a.click();});
window.addEventListener("pagehide",()=>{abort?.abort();if(objectURL)URL.revokeObjectURL(objectURL);});
prepare();
