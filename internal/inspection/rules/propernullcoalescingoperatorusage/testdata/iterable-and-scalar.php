<?php
namespace Doctrine\Common\Collections {
    /** @template T */
    interface Collection extends \IteratorAggregate, \Countable {}
}

namespace App {
    use Doctrine\Common\Collections\Collection;

    class Shelf implements \IteratorAggregate {
        public function getIterator(): \Iterator { return new \ArrayIterator([]); }
    }
    class Room {
        public Shelf $shelf;
    }
    class Order {
        /** @var Collection<int, string> */
        public Collection $lines;
        public ?int $id = null;
    }
    class Engine {}

    class Inventory {
        private ?int $cap = null;

        public function demo(?Room $room, ?\DOMElement $el, ?\Iterator $it, ?float $ratio, ?Order $order, ?Engine $engine, int|Engine|null $mixed) {
            foreach ($room?->shelf ?? [] as $item) {}
            $attrs = iterator_to_array($el?->attributes ?? []);
            $rows = $it ?? [];
            $shown = $ratio ?? '-';
            foreach ($order?->lines ?? [] as $l) {}
            $label = sprintf('#%s', $order?->id ?? 'new');
            $nodes = $el?->childNodes ?? [];
            return [
                $this->cap ?? 'none',
                $attrs, $rows, $shown, $label, $nodes,
                <weak_warning descr="Operand types of '??' do not match ([\App\Engine] vs [string]).">$engine ?? 'none'</weak_warning>,
                <weak_warning descr="Operand types of '??' do not match ([\App\Engine] vs [array]).">$engine ?? []</weak_warning>,
                <weak_warning descr="Operand types of '??' do not match ([\App\Engine, int] vs [string]).">$mixed ?? 'none'</weak_warning>,
                <weak_warning descr="Operand types of '??' do not match ([float] vs [array]).">$ratio ?? []</weak_warning>,
            ];
        }
    }
}
