<?php
foreach ($items as $item) {
    $made[$item->getId()] = new class(<warning descr="This call is repeated in the key; store its result in a local variable.">$item->getId()</warning>) {
        public function __construct(public $id) {}
    };
    $made[md5($item)] = [<warning descr="This call is repeated in the key; store its result in a local variable.">md5($item)</warning>, fn () => md5($item)];
}
