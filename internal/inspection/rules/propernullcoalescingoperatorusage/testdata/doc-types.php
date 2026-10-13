<?php
interface MockDouble {}
class Repository {}

class Factory
{
    /**
     * @template T of object
     * @param class-string<T> $type
     * @return MockDouble&T
     */
    public function double(string $type): MockDouble { throw new \LogicException(); }

    /**
     * @template K of string|null
     * @param K $key
     * @return (K is null ? array<string, mixed> : array<mixed>)
     */
    public function all(?string $key = null): array { return []; }

    /** @return iterable<int, string> */
    public function rows(): ?iterable { return null; }
}

function build(Factory $f, ?Repository $repo, ?Factory $other)
{
    // the double is a Repository too (MockDouble&T)
    $r = $repo ?? $f->double(Repository::class);
    // conditional return types resolve to their branches
    $all = $other?->all() ?? [];
    // iterable includes array
    foreach ($f->rows() ?? [] as $row) {}
    $bad = <weak_warning descr="Operand types of '??' do not match ([\Repository] vs [array]).">$repo ?? []</weak_warning>;
}
