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

function consent() {
    var uri = "/auth/consent";

    var ticket = location.hash.substring(1);
    var queries = {};
    var q = window.location.search.substring(1).split("&");
    for (var i = 0; i < q.length; i++) {
        var elem = q[i].split("=");

        var key = elem[0];
        var val = elem[1];
        if (val) {
            val = decodeURIComponent(val.replace(/\+/g, " "));
        }

        queries[key] = val;
    }

    document.write('ticket: ' + ticket + '<br/>');
    for (key in queries) {
        document.write(key + ': ' + queries[key] + '<br/>');
    }

    var scopes = "";
    if (queries["scope"]) {
        scopes = queries["scope"];
    }
    var claims = "";
    if (queries["claim"]) {
        claims = queries["claim"];
    }

    document.write('<form method="post" action="' + uri + '">');
    document.write('同意するスコープ: <input type="text" name="consented_scope" size="50" value="' + scopes + '" /><br/>');
    document.write('同意するクレーム: <input type="text" name="consented_claim" size="50" value="' + claims + '" /><br/>');
    document.write('拒否するスコープ: <input type="text" name="denied_scope" size="50" /><br/>');
    document.write('拒否するクレーム: <input type="text" name="denied_claim" size="50" /><br/>');
    document.write('<input type="hidden" name="ticket" value="'+ticket+'" />');
    document.write('<input type="submit" value="確認" />');
    document.write('</form>');
}
