<?php
class Sealer
{
    private \Random\Randomizer $rand;

    public function seal(string $data, string $key, \Random\Randomizer $given): array
    {
        $iv = (new \Random\Randomizer())->getBytes(16);
        $a = openssl_encrypt($data, 'aes-256-cbc', $key, OPENSSL_RAW_DATA, $iv);
        $b = openssl_encrypt($data, 'aes-256-cbc', $key, 0, (new \Random\Randomizer)->getBytes(16));

        $r = new \Random\Randomizer(new \Random\Engine\Secure());
        $c = openssl_encrypt($data, 'aes-256-cbc', $key, 0, $r->getBytes(16));
        $d = openssl_encrypt($data, 'aes-256-cbc', $key, 0, $this->rand->getBytes(16));
        $e = openssl_encrypt($data, 'aes-256-cbc', $key, 0, $given->GetBytes(16));
        return [$a, $b, $c, $d, $e];
    }

    public function weak(string $data, string $key, \Random\Engine $engine): array
    {
        $seeded = new \Random\Randomizer(new \Random\Engine\Mt19937(42));
        $a = openssl_encrypt($data, 'aes-256-cbc', $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: $seeded->getBytes(16).">$seeded->getBytes(16)</error>);
        $b = openssl_encrypt($data, 'aes-256-cbc', $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: (new \Random\Randomizer($engine))->getBytes(16).">(new \Random\Randomizer($engine))->getBytes(16)</error>);
        $c = openssl_encrypt($data, 'aes-256-cbc', $key, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: $this->source()->getBytes(16).">$this->source()->getBytes(16)</error>);
        return [$a, $b, $c];
    }

    private function source()
    {
        return $this;
    }
}
