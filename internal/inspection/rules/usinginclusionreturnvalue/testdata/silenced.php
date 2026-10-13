<?php
$config = @<weak_warning descr="Avoid relying on the value returned by an included file.">include 'config.php'</weak_warning>;
if (!@<weak_warning descr="Avoid relying on the value returned by an included file.">include 'optional.php'</weak_warning>) {
    fallback();
}
