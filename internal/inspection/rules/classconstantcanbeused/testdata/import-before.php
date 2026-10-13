<?php
namespace App;

class Item {}

// no existing import: the import goes before the first statement of the namespace
$engine = <weak_warning descr="Use \Random\Engine\Mt19937::class instead of the class name string.">'Random\Engine\Mt19937'</weak_warning>;
