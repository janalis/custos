<?php
class_exists($class, false) || require $file;
$loaded or include_once 'helpers.php';
$enabled && (require 'feature.php');
$cached ?? @include 'cache.php';
$ok = $loaded || <weak_warning descr="Avoid relying on the value returned by an included file.">require 'a.php'</weak_warning>;
(<weak_warning descr="Avoid relying on the value returned by an included file.">include 'b.php'</weak_warning>) || fallback();
$a . <weak_warning descr="Avoid relying on the value returned by an included file.">include 'c.php'</weak_warning>;
