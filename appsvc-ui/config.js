//overrides from bower_components/usergrid-portal/portal/dist/usergrid-portal/config.js
Usergrid = Usergrid || {};
if(!Usergrid || !Usergrid.options){
  console.error('missing other config reference');
}
// used only if hostname does not match a real server name
Usergrid.overrideUrl = 'https://api.usergrid.com/';
Usergrid.showNotifcations = true;
Usergrid.options.showAutoRefresh =true;
Usergrid.options.autoUpdateTimer=61;//seconds
Usergrid.options.client.requiresDeveloperKey = false;
// Usergrid.options.client.apiKey:'123456';
Usergrid.options.client.max = true; // change this for max menu
//Google Analytics
var _gaq = _gaq || [];
_gaq.push(['_setAccount', 'UA-4084158-4']);
