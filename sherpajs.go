package sherpa

var sherpaJS = []byte(`
(function(undefined) {
'use strict';

// prepare basic support for promises.
// we return functions with a "then" method only. our "then" isn't chainable. and you don't get other promise-related methods.
// but this "then" is enough so your browser's promise library (or a polyfill) can turn it into a real promise.
var thenable = function(fn) {
	var settled = false;
	var fulfilled = false;
	var result = null;

	var goods = [];
	var bads = [];

	// promise lib will call the returned function, make it the same as our .then function
	var nfn = function(goodfn, badfn) {
		if(settled) {
			if(fulfilled && goodfn) {
				goodfn(result);
			}
			if(!fulfilled && badfn) {
				badfn(result);
			}
		} else {
			if(goodfn) {
				goods.push(goodfn);
			}
			if(badfn) {
				bads.push(badfn);
			}
		}
	};
	nfn.then = nfn;

	var done = function() {
		while(fulfilled && goods.length > 0) {
			goods.shift()(result);
		}
		while(!fulfilled && bads.length > 0) {
			bads.shift()(result);
		}
	};

	var makeSettle = function(xfulfilled) {
		return function(arg) {
			if(settled) {
				return;
			}
			settled = true;
			fulfilled = xfulfilled;
			result = arg;
			done();
		};
	};
	var resolve = makeSettle(true);
	var reject = makeSettle(false);
	try {
		fn(resolve, reject);
	} catch(e) {
		reject(e);
	}
	return nfn;
};

var postJSON = function(url, param, success, error) {
	var req = new XMLHttpRequest();
	req.open('POST', url, true);
	req.onload = function() {
		if(req.status >= 200 && req.status < 400) {
			success(JSON.parse(req.responseText));
		} else {
			if(req.status === 404) {
				error({code: 'sherpaBadFunction', message: 'function does not exist'});
			} else {
				error({code: 'sherpaHttpError', message: 'error calling function, HTTP status: '+req.status});
			}
		}
	};
	req.onerror = function() {
		error({code: 'sherpaClientError', message: 'connection failed'});
	};
	req.setRequestHeader('Content-Type', 'application/json');
	req.send(JSON.stringify(param));
};

var make = function(api, name) {
	return function() {
		var params = Array.prototype.slice.call(arguments, 0);
		return api._wrapThenable(thenable(function(resolve, reject) {
			postJSON(api._sherpa.baseurl+name, {params: params}, function(response) {
				if(response && response.error) {
					reject(response.error);
				} else if(response && response.hasOwnProperty('result')) {
					resolve(response.result);
				} else {
					reject({code: 'sherpaBadResponse', message: "invalid sherpa response object, missing 'result'"});
				}
			}, reject);
		}));
	};
};

var api = {};
api._sherpa = SHERPA_JSON;

api._wrapThenable = function(thenable) {
	return thenable;
};

api._call = function(name) {
	return make(api, name).apply(Array.prototype.slice.call(arguments, 1));
};

for(var i = 0; i < api._sherpa.functions.length; i++) {
	var fn = api._sherpa.functions[i];
	api[fn] = make(api, fn);
}

window[api._sherpa.id] = api;

})();
`)
