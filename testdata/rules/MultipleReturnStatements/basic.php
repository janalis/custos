<?php
class Grader {
    public function <warning descr="3 return statements in this method; try to funnel them into a single exit.">letter</warning>(int $score) {
        if ($score > 90) {
            return 'A';
        }
        if ($score > 75) {
            return 'B';
        }
        return 'C';
    }

    public function <error descr="6 return statements in this method; try to funnel them into a single exit.">bucket</error>(int $n) {
        switch ($n) {
            case 1: return 'one';
            case 2: return 'two';
            case 3: return 'three';
            case 4: return 'four';
            case 5: return 'five';
        }
        return 'many';
    }

    public function sorter() {
        return function ($a, $b) {
            if ($a < $b) { return -1; }
            if ($a > $b) { return 1; }
            return 0;
        };
    }

    public function widget() {
        return new class {
            public function <warning descr="3 return statements in this method; try to funnel them into a single exit.">kind</warning>($v) {
                if (is_int($v)) { return 'int'; }
                if (is_string($v)) { return 'string'; }
                return 'other';
            }
        };
    }
}

abstract class Base {
    abstract public function nothing();
}

function plain($x) {
    if ($x) { return 1; }
    if (!$x) { return 2; }
    return 3;
}
