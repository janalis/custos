<?php
namespace App;

use Random\Engine\{Mt19937, Secure as Csprng};

class Item {
    public function name() {
        return Item::class;
    }
}

$engines = [
    Mt19937::class,
    Csprng::class,
];
