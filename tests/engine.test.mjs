import test from "node:test";
import assert from "node:assert/strict";
import {experiment,boundedUTF8} from "../web/engine.js";

const options={timeoutMS:10000,holdMS:1000,pollMS:2000,navigationRetryMS:5000};
function harness(overrides={}){
  let now=0,navigated=false,probes=0;
  const controller=new AbortController();
  const deps={now:()=>now,signal:controller.signal,closed:()=>false,sleep:async(ms)=>{now+=ms;},probe:async()=>{probes++;return {kind:"entry"};},navigate:()=>{navigated=true;},read:()=>({kind:navigated?"target":"stale",document:"proof",url:"http://test/proof"}),...overrides};
  return {deps,controller,probes:()=>probes};
}
test("rejected fetch alone never reports success",async()=>{const h=harness({probe:async()=>{throw new TypeError("Network error");},read:()=>({kind:"other"})});const r=await experiment(options,h.deps);assert.equal(r.outcome,"timed_out");assert.equal(r.saw_fetch_rejection,true);assert.equal(r.navigations,2);});
test("rejected fetch plus verified document reports navigation observation",async()=>{const h=harness({probe:async()=>{throw new TypeError("Failed to fetch");}});const r=await experiment(options,h.deps);assert.equal(r.outcome,"navigation_after_fetch_rejection");assert.equal(r.document,"proof");});
test("readable fetch does not masquerade as fetch-blocking bypass",async()=>{const h=harness({probe:async()=>({kind:"target"})});const r=await experiment(options,h.deps);assert.equal(r.outcome,"navigation_confirmed");assert.equal(r.fetch_proof,true);});
test("Host rejection yields protected result",async()=>{const h=harness({read:()=>({kind:"host_rejected"})});const r=await experiment(options,h.deps);assert.equal(r.outcome,"host_header_protected");});
test("popup closure is terminal",async()=>{const h=harness({closed:()=>true});const r=await experiment(options,h.deps);assert.equal(r.outcome,"popup_closed_or_severed");assert.equal(r.navigations,0);});
test("cross-origin errors do not confirm rebinding",async()=>{const h=harness({read:()=>{throw new DOMException("Cross origin","SecurityError");}});const r=await experiment(options,h.deps);assert.equal(r.outcome,"timed_out");});
test("cancellation terminates without extra navigation",async()=>{const h=harness();h.deps.probe=async()=>{h.controller.abort();throw new DOMException("Aborted","AbortError");};const r=await experiment(options,h.deps);assert.equal(r.outcome,"stopped");assert.equal(r.navigations,0);});
test("probes honor polling interval",async()=>{const h=harness({read:()=>({kind:"entry"})});await experiment(options,h.deps);assert.equal(h.probes(),5);});
test("navigation failure is bounded",async()=>{const h=harness({navigate:()=>{throw new Error("Navigation prevented");}});const r=await experiment(options,h.deps);assert.equal(r.outcome,"navigation_error");});
test("document cap counts UTF-8 bytes without splitting characters",()=>{
  for(const limit of [1,2,3,4,5,6,7,8]){const text=boundedUTF8("🔬éabcdef",limit);assert.ok(new TextEncoder().encode(text).length<=limit);assert.ok(!text.includes("�"));}
  assert.equal(boundedUTF8("abc",2),"ab");
});
