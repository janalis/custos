<?php
trait Odd {
    public $oct = 09;
}

class Broken {
    use Odd;
    public $oct = 9;
    public int ;
}
