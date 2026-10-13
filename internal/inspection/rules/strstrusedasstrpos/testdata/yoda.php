<?php
if (<warning descr="Use 'false === strpos($text, $word)' instead; it avoids building a substring.">!strstr($text, $word)</warning>) {}
$x = <warning descr="Use 'false !== stripos($text, $word)' instead; it avoids building a substring.">stristr($text, $word) !== false</warning>;
