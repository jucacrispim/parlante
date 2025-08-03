async function parlanteLoadComments(parlante_url, client_uuid, container_id) {
  let url = parlante_url + '/comment/' + client_uuid + '/html';
  let container = document.getElementById(container_id);
  let lang = navigator.language;
  let tz = Intl.DateTimeFormat().resolvedOptions().timeZone;

  let headers = new Headers();
  headers.append("Accepted-Language", lang)
  headers.append("X-Timezone", tz)

  let opts = {
    method: "GET",
    headers: headers,
    mode: "cors",
    cache: "no-cache",
    referrerPolicy: "unsafe-url",
  }

  let response = null
  try{
    response = await fetch(url, opts);
  }catch{
    container.innerHTML = 'Failed to load comments';
    return
  }

  let html = await response.text()
  container.innerHTML = html
  let btn = document.getElementById('parlante-submit')
  btn.onclick = function() {
    parlanteSubmitComment(parlante_url, client_uuid)
  }
}

async function parlanteSubmitComment(parlante_url, client_uuid) {
  let url = parlante_url + '/comment/' + client_uuid;
  let authorEl = document.getElementById("parlante-author")
  let contentEl = document.getElementById("parlante-content")
  let body = JSON.stringify({
    name: authorEl.value,
    content: contentEl.value,
  })

  let opts = {
    method: "POST",
    mode: "cors",
    cache: "no-cache",
    referrerPolicy: "unsafe-url",
    body: body,
  }

  let container_id = "parlante-add-comment"
  let container_ok_id = "parlante-add-ok"
  let container_error_id = "parlante-add-error"
  let container = document.getElementById(container_id)
  let container_ok = document.getElementById(container_ok_id)
  let container_error = document.getElementById(container_error_id)
  try{
    let response = await fetch(url, opts)
  }catch {
    container.style.display = 'none'
    container_error.style.display = 'block'
    return
  }
  container.style.display = 'none'
  container_ok.style.display = 'block'
}
