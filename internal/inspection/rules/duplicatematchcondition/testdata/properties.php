<?php
final class Scanner
{
    private bool $comment = false;
    private bool $anchor = false;
    private static bool $quoted = false;
    private array $flags = ['a' => false, 'b' => false];

    public function feed(string $c): string
    {
        return match (true) {
            $this->comment => 'comment',
            $this->anchor => 'anchor',
            self::$quoted => 'quoted',
            $this->flags['a'] => 'a',
            $this->flags['b'] => 'b',
            default => $this->plain($c),
        };
    }

    private function plain(string $c): string
    {
        $this->anchor = $c === '&';
        $this->comment = $c === '#';
        return 'plain';
    }
}
