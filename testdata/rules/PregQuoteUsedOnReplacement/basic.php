<?php
$s=preg_replace("/item/",<warning descr="Keep pattern escaping out of literal replacement text.">preg_quote("a.b","/")</warning>,$text);
