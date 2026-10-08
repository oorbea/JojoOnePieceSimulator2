package main

import (
	"html/template"
	"os"
)

// reviewTmpl is a self-contained page: radios per fruit (default "none" so a
// fruit is only uploaded after an explicit pick) and a button that downloads
// selections.json ({fruitID: candidateFile or ""}).
var reviewTmpl = template.Must(template.New("review").Parse(`<!doctype html>
<html lang="es"><head><meta charset="utf-8"><title>picturesync review</title>
<style>
body{font-family:system-ui,sans-serif;margin:0;background:#101418;color:#e8edf2}
header{position:sticky;top:0;background:#1b2229;padding:12px 16px;display:flex;gap:16px;align-items:center;z-index:1}
button{padding:8px 14px;font-size:15px;cursor:pointer}
section{padding:12px 16px;border-bottom:1px solid #2a343d}
h2{margin:0 0 8px;font-size:16px}h2 small{color:#8fa1b3;font-weight:normal}
.row{display:flex;gap:12px;flex-wrap:wrap}
label{display:flex;flex-direction:column;align-items:center;gap:4px;border:2px solid transparent;padding:6px;border-radius:8px;cursor:pointer;font-size:13px}
label:has(input:checked){border-color:#4aa3ff;background:#1b2a3a}
img{width:200px;height:200px;object-fit:contain;background:#0a0d10}
.none{justify-content:center;width:120px;color:#8fa1b3}
</style></head><body>
<header><button id="export">Exportar selections.json</button><span id="count"></span></header>
{{range .}}{{$id := .ID}}<section data-id="{{.ID}}">
<h2>{{.Name}} <small>{{.Rarity}}</small></h2>
<div class="row">
<label class="none"><input type="radio" name="{{.ID}}" value="" checked> ninguna</label>
{{range .Candidates}}<label><img src="{{.File}}" loading="lazy" alt=""><span><input type="radio" name="{{$id}}" value="{{.File}}"> {{.Page}}</span></label>
{{end}}</div></section>
{{end}}
<script>
const count=()=>{document.getElementById('count').textContent=
  document.querySelectorAll('input:checked:not([value=""])').length+' seleccionadas'};
document.addEventListener('change',count);count();
document.getElementById('export').onclick=()=>{
  const sel={};
  document.querySelectorAll('section').forEach(s=>{
    const c=s.querySelector('input:checked');sel[s.dataset.id]=c?c.value:'';});
  const a=document.createElement('a');
  a.href=URL.createObjectURL(new Blob([JSON.stringify(sel,null,2)],{type:'application/json'}));
  a.download='selections.json';a.click();};
</script></body></html>
`))

func writeReview(path string, entries []candidateEntry) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return reviewTmpl.Execute(f, entries)
}
