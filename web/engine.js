// Browser-independent state machine. Dependencies isolate network/browser
// behavior so errors, cancellation and evidence handling can be tested.
export function boundedUTF8(text, limit) {
  const prefix=text.slice(0,limit), bytes=new TextEncoder().encode(prefix);
  if(bytes.length<=limit)return prefix;
  // Streaming decode omits a trailing partial multibyte sequence.
  return new TextDecoder().decode(bytes.subarray(0,limit),{stream:true});
}
export async function experiment(options, deps) {
  const start=deps.now(), deadline=start+options.timeoutMS;
  let lastNavigation=-Infinity, nextProbe=start, sawFetchRejection=false, fetchProof=false, navigations=0;
  const observations=[];
  const emit=(kind,detail)=>{const event={at_ms:Math.round(deps.now()-start),kind,detail};observations.push(event);deps.event?.(event);};
  const finish=(outcome,evidence={})=>({outcome,elapsed_ms:Math.round(deps.now()-start),saw_fetch_rejection:sawFetchRejection,fetch_proof:fetchProof,navigations,observations,...evidence});
  emit("waiting","DNS switch is scheduled; browser caches may retain the entry address.");
  while(deps.now()<deadline){
    if(deps.signal?.aborted){emit("stopped","Cancelled by the operator.");return finish("stopped");}
    if(deps.closed()){emit("popup_closed","Popup closed or its reference was severed.");return finish("popup_closed_or_severed");}
    if(navigations>0){
      let view;
      try{view=deps.read();}catch(err){view={kind:"unreadable",error:err.name||"Error"};}
      if(view?.kind==="target"){
        emit("navigation_confirmed","Read the target proof marker through the popup window reference.");
        return finish(sawFetchRejection&&!fetchProof?"navigation_after_fetch_rejection":"navigation_confirmed",{document:view.document,url:view.url,truncated:!!view.truncated});
      }
      if(view?.kind==="host_rejected"){emit("host_protected","Fixture rejected the rebinding Host header.");return finish("host_header_protected");}
    }
    const elapsed=deps.now()-start;
    if(deps.now()>=nextProbe){
      // Fetch is a comparison channel, never proof of the reason for blocking.
      try{
        const probe=await deps.probe(Math.max(1,deadline-deps.now()));
        if(deps.signal?.aborted)continue;
        if(probe.kind==="target"&&!fetchProof){fetchProof=true;emit("fetch_target_readable","The target is readable through fetch; this run does not demonstrate a fetch-blocking bypass.");}
      }catch(err){
        if(deps.signal?.aborted)continue;
        if(!sawFetchRejection){emit("fetch_rejected",`${err.name||"Error"}: ${String(err.message||err).slice(0,250)}. Cause is not established by JavaScript.`);}
        sawFetchRejection=true;
      } finally {nextProbe=deps.now()+options.pollMS;}
    }
    if(deps.signal?.aborted)continue;
    if(deps.now()>=deadline)break;
    if((sawFetchRejection||elapsed>=options.holdMS)&&deps.now()-lastNavigation>=options.navigationRetryMS){
      try{deps.navigate();navigations++;lastNavigation=deps.now();emit("navigation_attempt",`Top-level popup navigation ${navigations}.`);}
      catch(err){emit("navigation_error",String(err.message||err).slice(0,250));return finish("navigation_error");}
    }
    try{await deps.sleep(Math.min(250,Math.max(1,deadline-deps.now())),deps.signal);}catch(err){if(!deps.signal?.aborted)throw err;}
  }
  emit("timed_out","No target proof confirmed before the run deadline. Inspect DNS, popup errors, browser permissions and target headers.");
  return finish("timed_out");
}
