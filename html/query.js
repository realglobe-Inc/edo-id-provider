// Copyright 2015 realglobe, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

function query_parse(raw) {
    var queries = {};
    var q = raw.split("&");
    for (var i = 0; i < q.length; i++) {
        var elem = q[i].split("=");

        var key = elem[0];
        var val = elem[1];
        if (val) {
            val = decodeURIComponent(val.replace(/\+/g, " "));
        }

        queries[key] = val;
    }

    return queries;
}

function display() {
    var ticket = location.hash.substring(1);
    var queries = query_parse(window.location.search.substring(1));

    if (ticket) {
        document.write('ticket: ' + ticket + '<br/>');
    }
    if (queries && Object.keys(queries).length > 0) {
        for (key in queries) {
            document.write(key + ': ' + queries[key] + '<br/>');
        }
    }
}
