<?php
namespace Lib;

class Remote
{
    public function same($a = [1, 2]) { return $a; }
    public function other($a = 1) { return $a; }
    /** @return void */
    public function quiet($a) { echo $a; }
    public function loud($a) { echo $a; }
    public function documented($a) { return $a; }
}
