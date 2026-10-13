<?php
function render(string $region, array $links): array
{
    $page = ['#markup' => 'Rows'];
    $page['#attached']['library'][] = 'shop/cart';
    $page[$region][] = $links;
    $page['meta']['author'] = 'me';
    return $page;
}

class Paths
{
    /** @var string[] */
    protected $paths = [];

    /** @var string[] */
    protected static $names = [];

    public function add(string $ns, string $dir): void
    {
        <error descr="Appending with [] is not supported on strings (fatal error).">$this->paths[$ns][]</error> = $dir;
        <error descr="Appending with [] is not supported on strings (fatal error).">self::$names[$ns][]</error> = $dir;
        $fn = function () use ($ns) {
            $local = ['a' => 'b'];
            $local['z'][] = $ns;
        };
    }
}
