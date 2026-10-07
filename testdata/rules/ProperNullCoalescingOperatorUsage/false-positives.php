<?php
class Moment extends \DateTimeImmutable
{
    private ?int $code = null;
    /** @var array<string, string> */
    private array $labels = [];

    public static function now(): \DateTimeImmutable { return new \DateTimeImmutable(); }

    public function pick(?parent $reference = null, ?self $other = null, string $key = '')
    {
        $a = $reference ?? self::now();
        $b = $other ?? new Moment();
        $c = $this->code ?? throw new \LogicException('not ready');
        $d = $this->labels[$key] ?? throw new \LogicException('unknown');
        return [$a, $b, $c, $d];
    }
}

/** @param array|string[] $rows */
function rows($rows, array $raw, bool $flag) {
    $entry = $flag ? $raw : ['message' => $raw];
    return [$rows[0] ?? '', $entry['message'] ?? ''];
}
