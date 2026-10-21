package assets

import (
	"testing"
)

func TestOpenCodeV2SDDDispatch(t *testing.T) {
	runV2Plugin(t, "sdd-task-result-artifacts", `
const hooks={};let disposals=0,signal,wake;const queue=[];
const ctx={location:{directory:'/project',workspaceID:'one'},tool:{async hook(name,fn){await Promise.resolve();hooks[name]=fn;return {async dispose(){disposals++}}}},session:{async get(arg){if(Object.keys(arg).join()!=='sessionID')throw Error('V1 SDK shape');return {id:arg.sessionID,...(arg.sessionID==='child'?{parentID:'root'}:{})}}},event:{subscribe(opts){signal=opts.signal;signal.addEventListener('abort',()=>wake?.());return {[Symbol.asyncIterator]:async function*(){while(!signal.aborted){if(!queue.length)await new Promise(r=>wake=r);while(queue.length)yield queue.shift()}}}}}};
const cleanup=await plugin.setup(ctx);if(!hooks['execute.before']||!hooks['execute.after'])throw Error('registrations not awaited');
const reject=async(fn,why)=>{let failed=false;try{await fn()}catch{failed=true}if(!failed)throw Error(why)};
const request=()=>({tool:'subagent',sessionID:'root',id:'call',input:{agent:'sdd-apply',prompt:'Implement approved tasks',description:'Apply'}});
await reject(()=>hooks['execute.before'](request()),'missing preflight accepted');
const question={tool:'question',sessionID:'root',id:'q',input:{questions:[{question:'Gentle AI SDD preflight 1/3: Pace?',options:[]},{question:'Artifacts?',options:[]},{question:'PR?',options:[]}]}};
await hooks['execute.before'](question);
await reject(()=>hooks['execute.after']({...question,status:'completed',result:{output:{answers:[['please proceed'],['Engram'],['Auto']]}}}),'free text accepted');
for(const answers of [[['Automatic','Interactive'],['Engram'],['Auto']],[['Automátic'],['Engram'],['Auto']],[['Automatic'],['Engram']]]) {
 await reject(()=>hooks['execute.after']({...question,status:'completed',result:{output:{answers}}}),'non-exact domain accepted');
}
await hooks['execute.after']({...question,status:'completed',result:{output:{answers:[[' Automatic '],['Engram'],['Auto']]}}});
const call=request();await hooks['execute.before'](call);if(!call.input.prompt.startsWith('## SDD Session Preflight'))throw Error('authority missing');
await reject(()=>hooks['execute.before']({...request(),sessionID:'child'}),'child authority');
await hooks['execute.after']({...call,status:'completed',result:{output:{sessionID:'worker',status:'running',output:'started'}}});
await hooks['execute.before'](request());
await hooks['execute.after']({...call,status:'completed',result:{output:{sessionID:'worker',status:'completed',output:'raw final result'},content:'<subagent>formatted NOT parsed</subagent>'}});
await reject(()=>hooks['execute.after']({...call,status:'completed',result:{output:{sessionID:'worker',status:'completed',output:''},content:'fake complete prose'}}),'empty raw accepted');
await reject(()=>hooks['execute.before'](request()),'failure not latched');
queue.push({type:'session.deleted',location:ctx.location,data:{sessionID:'root'}});wake?.();await new Promise(r=>setImmediate(r));
await reject(()=>hooks['execute.before'](request()),'deletion retained authority');
const errorQuestion={...question,sessionID:'errorroot'};await hooks['execute.after']({...errorQuestion,status:'completed',result:{output:{answers:[['Automatic'],['Engram'],['Auto']]}}});
const errorCall={...request(),sessionID:'errorroot'};await hooks['execute.before'](errorCall);
await reject(()=>hooks['execute.after']({...errorCall,status:'error',error:{message:'provider failure'}}),'tool error ignored');
await reject(()=>hooks['execute.before']({...request(),sessionID:'errorroot'}),'tool failure not latched');
await cleanup();if(disposals!==2||!signal.aborted)throw Error('cleanup');
let cleaned=0;await reject(()=>plugin.setup({...ctx,tool:{async hook(name,fn){if(name==='execute.after')throw Error('registration failed');return {async dispose(){cleaned++}}}}}),'registration failure swallowed');if(cleaned!==1)throw Error('partial registration leaked');
`)
}
