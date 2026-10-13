<?php

namespace Shop\Flow;

final class Basket implements \Countable
{
    public function count(): int { return 0; }
}

final class Registry implements \ArrayAccess
{
    public function offsetExists(mixed $offset): bool { return false; }
    public function offsetGet(mixed $offset): mixed { return null; }
    public function offsetSet(mixed $offset, mixed $value): void {}
    public function offsetUnset(mixed $offset): void {}
}

final class Title implements \Stringable
{
    public function __toString(): string { return 'title'; }
}

final class Stage
{
    public function callables(array $rows, ?callable $keep = null, ?callable $shape = null, ?\Closure $late = null, ?string $name = null, ?callable $last = null)
    {
        $keep = $keep ?? 'is_string';
        $shape = $shape ?? [$this, 'same'];
        $late = <weak_warning descr="Operand types of '??' do not match ([callable] vs [string]).">$late ?? 'strtoupper'</weak_warning>;
        $named = $name ?? $shape;
        $bad = <weak_warning descr="Operand types of '??' do not match ([callable] vs [int]).">$last ?? 0</weak_warning>;
        return [array_map($shape, array_filter($rows, $keep)), $late, $named, $bad];
    }

    public function same($value) { return $value; }

    public function arrayLike(?\Countable $any, ?Basket $basket, ?\ArrayAccess $bag, ?Registry $registry, array $plain = [])
    {
        $size = count($any ?? []) + count($basket ?? []);
        $store = $bag ?? [];
        $other = $registry ?? [];
        $back = $plain[0] ?? $registry;
        $bad = <weak_warning descr="Operand types of '??' do not match ([\Shop\Flow\Basket] vs [string]).">$basket ?? ''</weak_warning>;
        $mixed = <weak_warning descr="Operand types of '??' do not match ([\Countable] vs [\Shop\Flow\Registry]).">$any ?? new Registry()</weak_warning>;
        return [$size, $store['k'] ?? null, $other, $back, $bad, $mixed];
    }

    public function stringLike(?\Stringable $label, ?Title $title, ?string $text = null)
    {
        $shown = 'Hi ' . ($label ?? '');
        $kept = $text ?? $label;
        $bad = <weak_warning descr="Operand types of '??' do not match ([\Shop\Flow\Title] vs [string]).">$title ?? 'untitled'</weak_warning>;
        $list = <weak_warning descr="Operand types of '??' do not match ([\Stringable] vs [array]).">$label ?? []</weak_warning>;
        return [$shown, $kept, $bad, $list];
    }
}
