<?php
namespace Shop {
    use Shop\Billing\Invoice as Bill;

    interface Pricing {
        /** @return int|null */
        function <weak_warning descr="Declare ': ?int' as the return type.">discount</weak_warning>();
        function undocumented();
    }

    final class Cart {
        public function __toString()
        {
            return 'cart';
        }

        /** @param float $p */
        public function <weak_warning descr="Declare ': float' as the return type.">price</weak_warning>($p) { return $p; }

        public function <weak_warning descr="Declare ': void' as the return type.">clear</weak_warning>() { $this->n = 0; }

        public function <weak_warning descr="Declare ': ?\ArrayObject' as the return type.">maybe</weak_warning>($f) {
            if ($f) {
                return new \ArrayObject();
            }
        }

        public function <weak_warning descr="Declare ': Bill' as the return type.">bill</weak_warning>() { return new \Shop\Billing\Invoice(); }

        public function <weak_warning descr="Declare ': Billing\Receipt' as the return type.">receipt</weak_warning>() { return new \Shop\Billing\Receipt(); }

        /** @return $this */
        public function <weak_warning descr="Declare ': self' as the return type.">touch</weak_warning>() { return $this; }

        /** @param mixed $v */
        public function raw($v) { return $v; }

        public function both($f) { return $f ? 1 : 'one'; }

        public function <weak_warning descr="Declare ': \Generator' as the return type.">chars</weak_warning>($s) { yield $s; return 1; }
    }
}

namespace Shop\Billing {
    class Invoice {}
    class Receipt {}
}
