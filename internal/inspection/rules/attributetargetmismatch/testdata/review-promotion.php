<?php
#[Attribute(Attribute::TARGET_PROPERTY)]
final class Column {}

final class Row
{
    public function __construct(
        #[Column] readonly int $id,
        #[Column] private(set) string $name,
        #[Column] protected(set) string $label,
        #[Column] public(set) string $title,
        #[Column] final int $rank,
        #[Column] protected int $kept,
        #[Column] private int $hidden,
        #[Column] public readonly int $both,
        <error descr="Use the attribute on a permitted target.">#[Column]</error> int $plain,
    ) {}

    public function rename(<error descr="Use the attribute on a permitted target.">#[Column]</error> string $name): void {}
}
