<?php
namespace Routing;

function preg_quote($text) { return \preg_quote($text, '#'); }

$a = '#' . preg_quote($segment) . '#';
$b = '#' . \Text\preg_quote($segment) . '#';
$c = '~' . \<error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($segment) . '~';
