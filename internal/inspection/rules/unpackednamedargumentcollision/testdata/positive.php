<?php
function send($id) {} send(...['id'=>7],<error descr="Bind each argument only once.">id:8</error>);
function three($a = 'A', $b = 'B', $c = 'C') {}
function rest($a = 'A', ...$r) {}
three(...['b' => 1], ...[7], <error descr="Bind each argument only once.">c: 3</error>);
three(...['a' => 1], ...[7], <error descr="Bind each argument only once.">a: 3</error>);
rest(...['q' => 1], ...[2], <error descr="Bind each argument only once.">a: 3</error>);
rest(...['q' => 1], <error descr="Bind each argument only once.">q: 2</error>);
