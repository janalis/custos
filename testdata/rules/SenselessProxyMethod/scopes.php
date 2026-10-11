<?php
namespace First {
    use Lib\Remote as Table;

    class Base
    {
        public function size($s = Table::WIDTH) { return $s; }
    }
}

namespace Second {
    // `Table` is not imported in this block, so the same text names another class.
    class Child extends \First\Base
    {
        public function size($s = Table::WIDTH) { return parent::size($s); }
    }
}
