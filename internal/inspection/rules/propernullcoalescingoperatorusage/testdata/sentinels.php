<?php
namespace App;

class Engine {}

interface Channel {
    public function ack(int $tag): void;
    public function stats(): ?array;
}

class Cache {
    private static ?array $allowed = null;

    /** @param array<string, Engine> $seen */
    public function check(array $seen, string $id, Channel $channel, int $tag): array {
        return [
            // A boolean fallback marks an absent value.
            $seen[$id] ?? false,
            !(self::$allowed ?? true),
            $channel->stats() ?? false,
            // A void call evaluates to null.
            $channel->ack($tag) ?? true,
            $channel->ack($tag) ?? 'done',
            <weak_warning descr="Operand types of '??' do not match ([array] vs [string]).">$channel->stats() ?? 'none'</weak_warning>,
        ];
    }
}
