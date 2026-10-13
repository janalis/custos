<?php
namespace Shop;

use Psr\Log\LoggerInterface;

class Money {}

class Wallet {
    public function add(Money $m): void {}
}

class <weak_warning descr="Depends on 3 distinct classes; consider splitting it up.">Checkout</weak_warning> extends Wallet {
    /** @var \Some\DocOnly */
    private $log;
    public function __construct(LoggerInterface $log, int $tries) { $this->log = $log; }
    public function pay(Money $m): self { return $this; }
}

class Basket {
    public function total(int $cents): string { return (string) $cents; }
}

#[\Attr\Marker]
enum Suit: string {
    case Hearts = 'H';
    public function x(): static { return strlen(FOO) ? $this : self::Hearts; }
}

interface <weak_warning descr="Depends on 2 distinct classes; consider splitting it up.">Port</weak_warning> extends \Countable {
    public function make(): ?Money;
}

trait <weak_warning descr="Depends on 6 distinct classes; consider splitting it up.">Helper</weak_warning> {
    public function h($x) {
        try {
            $y = new \ArrayObject();
            return $x instanceof Money || Wallet::class === $x || \Other\Cfg::$v;
        } catch (\RuntimeException | Money $e) {
            return new class extends Basket {};
        }
    }
}
