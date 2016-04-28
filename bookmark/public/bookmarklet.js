// jshint asi:true
// jshint unused:true
//
// http://chriszarate.github.io/bookmarkleter/
var xhr = new XMLHttpRequest()
xhr.open('POST', 'DOMAIN/api/bookmarks', true)
xhr.setRequestHeader("Content-Type", "application/json")
xhr.onerror = function() { alert('cannot make CORS request') }
xhr.onload = function() {
  console.log('completed', xhr)
}
xhr.send(JSON.stringify({url: location.href}))
