<?php
final class Session {
    public function __construct(
        private ?string $token = null,
        public readonly array $flags = [],
    ) {}

    public function active(): bool {
        return $this->token !== null;
    }

    public function missing(): bool {
        $none = $this->token === null;
        return $none;
    }
}
