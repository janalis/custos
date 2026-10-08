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
            $this->limit,
            $page,
            strtoupper('x'),
        ];
    }

    /** @return int */
    private function offset() { return 0; }

    /** @return string */
    private function name(): mixed { return ''; }
}
