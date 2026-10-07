<?php
namespace Shop {
    use Shop\Billing\Invoice as Bill;

    interface Pricing {
        /** @return int|null */
        function discount(): ?int;
        function undocumented();
    }

    final class Cart {
        public function __toString()
        {
            return 'cart';
        }

        /** @param float $p */
        public function price($p): float { return $p; }

        public function clear(): void { $this->n = 0; }

        public function maybe($f): ?\ArrayObject {
            if ($f) {
                return new \ArrayObject();
            }
        }

        public function bill(): Bill { return new \Shop\Billing\Invoice(); }

        public function receipt(): Billing\Receipt { return new \Shop\Billing\Receipt(); }

        /** @return $this */
        public function touch(): self { return $this; }

        /** @param mixed $v */
        public function raw($v) { return $v; }

        public function both($f) { return $f ? 1 : 'one'; }

        public function chars($s): \Generator { yield $s; return 1; }
    }
}

namespace Shop\Billing {
    class Invoice {}
    class Receipt {}
}
