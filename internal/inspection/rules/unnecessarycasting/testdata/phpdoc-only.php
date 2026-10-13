<?php
final class Query
{
    private int $limit = 10;

    /** @param array<int, int> $ids */
    public function build(array $ids, int $page): array
    {
        $parts = [];
        foreach ($ids as $id) {
            $parts[] = (int) $id; // only the docblock says int
        }
        /** @var int $offset */
        $offset = $this->offset();
        return [
            $parts,
            (int) $offset,
            (string) $this->name(),
            <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $this->limit,
            <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $page,
            <weak_warning descr="Operand already has the target type; remove the cast.">(string)</weak_warning> strtoupper('x'),
        ];
    }

    /** @return int */
    private function offset() { return 0; }

    /** @return string */
    private function name(): mixed { return ''; }
}
