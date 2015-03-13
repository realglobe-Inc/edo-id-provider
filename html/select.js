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

function select() {
    var uri = "/auth/select";

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

    var username = ""
    if (queries["usernames"]) {
        var buff = JSON.parse(queries["usernames"]);
        if (buff[0]) {
            username = buff[0];
        }
    }

    document.write('<form method="post" action="' + uri + '">');
    document.write('アカウント: <input type="text" name="username" size="20" value="' + username + '" /> ');
    document.write('<input type="hidden" name="ticket" value="' + ticket + '" />');
    document.write('<input type="submit" value="選択" />');
    document.write('</form>');
}
