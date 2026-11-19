<?php

/** @param array{id: int, name: string} $row */
function takesRow(array $row) {}

/** @param object{id: int} $item */
function takesItem(object $item) {}

/** @param callable(int): bool $test */
function takesTest(callable $test) {}

/**
 * @param array{id: string, name: string} $wrongValue
 * @param array{id: int} $missingKey
 * @param array{id: int, name: string, extra: int} $extraKey
 * @param list<int> $list
 * @param object{id: string} $wrongProperty
 * @param callable(int): array $wrongReturn
 * @param callable(array): bool $wrongParameter
 * @param callable(int, int): bool $tooManyRequired
 */
function mismatches(array $wrongValue, array $missingKey, array $extraKey, array $list, object $wrongProperty, callable $wrongReturn, callable $wrongParameter, callable $tooManyRequired)
{
    takesRow(<warning descr="Pass array{id:int,name:string} to $row of takesRow(), not array{id:string,name:string}.">$wrongValue</warning>);
    takesRow(<warning descr="Pass array{id:int,name:string} to $row of takesRow(), not array{id:int}.">$missingKey</warning>);
    takesRow(<warning descr="Pass array{id:int,name:string} to $row of takesRow(), not array{id:int,name:string,extra:int}.">$extraKey</warning>);
    takesRow(<warning descr="Pass array{id:int,name:string} to $row of takesRow(), not int[].">$list</warning>);
    takesItem(<warning descr="Pass object{id:int} to $item of takesItem(), not object{id:string}.">$wrongProperty</warning>);
    takesTest(<warning descr="Pass callable(int): (bool) to $test of takesTest(), not callable(int): (array).">$wrongReturn</warning>);
    takesTest(<warning descr="Pass callable(int): (bool) to $test of takesTest(), not callable(array): (bool).">$wrongParameter</warning>);
    takesTest(<warning descr="Pass callable(int): (bool) to $test of takesTest(), not callable(int, int): (bool).">$tooManyRequired</warning>);
}

/**
 * @param array{id: int, name: string} $row
 * @param array{id: int, name: string, extra?: int, ...} $open
 * @param array<string, mixed> $map
 * @param object{id: int, more: string} $item
 * @param object $plain
 * @param callable(int): bool $test
 * @param callable(int, int=): bool $optional
 * @param callable(int|string): bool $wider
 * @param callable $any
 */
function fits(array $row, array $open, array $map, object $item, object $plain, callable $test, callable $optional, callable $wider, callable $any)
{
    takesRow($row);
    takesRow($open);
    takesRow($map);
    takesItem($item);
    takesItem($plain);
    takesTest($test);
    takesTest($optional);
    takesTest($wider);
    takesTest($any);
}
