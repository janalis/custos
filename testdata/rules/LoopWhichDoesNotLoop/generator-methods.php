<?php
class Shelf {
    public function books() { return ['a', 'b']; }
    public function makers() { $make = function () { yield 1; }; return [$make]; }
    public function firstBook() {
        <warning descr="Loop body exits on the first iteration; the loop never repeats.">foreach</warning> ($this->books() as $book) {
            return $book;
        }
        <warning descr="Loop body exits on the first iteration; the loop never repeats.">foreach</warning> ($this->makers() as $maker) {
            return $maker;
        }
        <warning descr="Loop body exits on the first iteration; the loop never repeats.">foreach</warning> ($this->$dynamic() as $item) {
            break;
        }
        <warning descr="Loop body exits on the first iteration; the loop never repeats.">foreach</warning> ($this->missing() as $item) {
            break;
        }
        return null;
    }
}
