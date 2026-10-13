<?php
// error recovery: the parameter list is missing
interface Broken {
    /** @return int */
    public function <weak_warning descr="Declare ': int' as the return type.">size</weak_warning>;
}
