<?php
namespace App;

use Random\Engine\{Mt19937, Secure as Csprng};

class Item {
    public function name() {
        return <weak_warning descr="Use Item::class instead of the class name string.">__NAMESPACE__ . '\Item'</weak_warning>;
    }
}

$engines = [
    <weak_warning descr="Use \Random\Engine\Mt19937::class instead of the class name string.">'Random\Engine\Mt19937'</weak_warning>,
    <weak_warning descr="Use \Random\Engine\Secure::class instead of the class name string.">'\Random\Engine\Secure'</weak_warning>,
];
