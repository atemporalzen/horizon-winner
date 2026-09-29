const $ = (id) => document.getElementById(id);
let token = "", targets = [];
async function api(path, options = {}) {
  const r = await fetch(path, {...options, headers: {Authorization: `Bearer ${token}`, ...options.headers}});
  if (!r.ok) throw new Error(`Controller returned ${r.status}: ${await r.text()}`);
  return r.json();
}
$("connect").addEventListener("submit", async (e) => {
  e.preventDefault(); token = $("token").value; $("connect-status").textContent = "Connecting…";
  try {
    const c = await api("/api/config"); targets = c.targets;
    $("target").replaceChildren(...targets.map(t => {const o = document.createElement("option"); o.value=t.id; o.textContent=t.label; return o;}));
    $("token").value=""; $("experiment").hidden=false; $("connect-status").textContent="Connected. Token is held in this page's memory only."; detail();
  } catch (err) {token=""; $("experiment").hidden=true; $("connect-status").textContent=err.message;}
});
function detail() {const t=targets.find(t=>t.id===$("target").value); if(t) $("target-detail").textContent=`${t.ip}:${t.port}${t.path} · confirmation requires the configured proof marker`;}
$("target").addEventListener("change",detail);
$("create").addEventListener("submit",async(e)=>{
  e.preventDefault(); const button=e.currentTarget.querySelector("button"); button.disabled=true; $("session").hidden=true;
  try {const s=await api("/api/sessions",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({target_id:$("target").value})}); $("run-link").href=s.url; $("session").hidden=false; $("create-status").textContent=`Session ${s.id.slice(0,8)} · expires ${new Date(s.expires_at).toLocaleTimeString()}`;}
  catch(err){$("create-status").textContent=err.message;} finally{button.disabled=false;}
});
