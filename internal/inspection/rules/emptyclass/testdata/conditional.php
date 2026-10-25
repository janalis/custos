<?php
namespace App\Cache;

if (version_compare(phpversion('redis'), '6.2.0', '>=')) {
    trait ProxyTrait
    {
        public function getdel($key) { return $key; }
    }
} else {
    trait ProxyTrait
    {
    }
}

if (!class_exists(Fallback::class)) {
    class Fallback {}
}

class <weak_warning descr="This class declares no members; remove it or give it a purpose.">Unconditional</weak_warning> {}
