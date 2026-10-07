<?php
class T {
    public function <warning descr="2 return statements in this method; try to funnel them into a single exit.">two</warning>($x) {
        if ($x) { return 1; }
        return 2;
    }
    public function <error descr="3 return statements in this method; try to funnel them into a single exit.">three</error>($x) {
        if ($x) { return 1; }
        if ($x > 1) { return 3; }
        return 2;
    }
    public function one() { return 1; }
}
