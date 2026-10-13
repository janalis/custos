<?php

class Mt19937 {}

// importing Random\Engine\Mt19937 would redeclare the class above
$engine = <weak_warning descr="Use \Random\Engine\Mt19937::class instead of the class name string.">'Random\Engine\Mt19937'</weak_warning>;
