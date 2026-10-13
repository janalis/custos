<?php
namespace App\Returns;

final class Lookup {
    // An untyped parameter may hold anything: ': string' would throw.
    public function first($candidate) {
        if ($candidate) {
            return $candidate;
        }
        return '';
    }

    // Unresolvable call: its result is unknown.
    public function bump($key) {
        if ($key === '') {
            return false;
        }
        return \not_declared_anywhere($key);
    }

    // A @return tag describes the unknown values: still reported.
    /** @return string */
    public function named($candidate) {
        if ($candidate) {
            return $candidate;
        }
        return '';
    }
}
