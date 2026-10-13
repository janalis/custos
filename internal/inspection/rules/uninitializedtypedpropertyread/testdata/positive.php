<?php
class Meter {public int $value;} echo <error descr="Initialize this typed property before reading it.">(new Meter())->value</error>;

class Conf {public array $opts; public ?Conf $next; public int $n; public function items(): array {return [];}}
isset(<error descr="Initialize this typed property before reading it.">(new Conf)->next</error>->items()['k']);
echo <error descr="Initialize this typed property before reading it.">(new Conf)->next</error>->items()['k'] ?? 'fallback';
echo [1][<error descr="Initialize this typed property before reading it.">(new Conf)->n</error>] ?? 'fallback';
echo null ?? <error descr="Initialize this typed property before reading it.">(new Conf)->opts</error>['k'];
echo <error descr="Initialize this typed property before reading it.">(new Conf)->opts</error>['k'] + 1;
final class Sealed {public int $v; public function peek() {return <error descr="Initialize this typed property before reading it.">(new static)->v</error>;}}
