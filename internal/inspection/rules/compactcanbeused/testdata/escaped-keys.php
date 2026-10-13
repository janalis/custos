<?php
function view($tag, $title, $body) {
    $a = <weak_warning descr="Replace with 'compact(&quot;t\x61g&quot;, 'title')'.">[</weak_warning>"t\x61g" => $tag, 'title' => $title];
    $b = <weak_warning descr="Replace with 'compact(&quot;b\157dy&quot;, &quot;\164itle&quot;)'.">array</weak_warning>("b\157dy" => $body, "\164itle" => $title);
    $c = ["t\x61gs" => $tag, 'title' => $title];
    $d = ['t\x61g' => $tag, 'title' => $title];
    return [$a, $b, $c, $d];
}
