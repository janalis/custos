<?php
final class Session {
    public function __construct(
        private ?string $token = null,
        public readonly array $flags = [],
    ) {}

    public function active(): bool {
        return <weak_warning descr="Compare with null instead: '$this->token !== null'.">isset($this->token)</weak_warning>;
    }

    public function missing(): bool {
        $none = <weak_warning descr="Compare with null instead: '$this->token === null'.">!isset($this->token)</weak_warning>;
        return $none;
    }
}
