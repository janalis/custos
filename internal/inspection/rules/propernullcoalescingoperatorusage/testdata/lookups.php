<?php
final class Node { public float $marginX = 0.0; }
final class Cluster { public float $marginX = 1.0; }

final class Layout
{
    private ?Node $node = null;
    private Cluster $cluster;

    /**
     * @param array<string, Node> $nodes
     * @param array<string, Cluster> $groups
     * @param array<string, int> $counts
     */
    public function margin(array $nodes, array $groups, string $id, ?Node $n, Cluster $c, int $i, array $counts): float
    {
        $owner = $nodes[$id] ?? $groups[$id];
        $owner = <weak_warning descr="Operand types of '??' do not match ([\Node] vs [\Cluster]).">$n ?? $c</weak_warning>;
        $owner = ($nodes[$id]) ?? $groups[$id];
        $owner = <weak_warning descr="Operand types of '??' do not match ([\Node] vs [\Cluster]).">$n ?? new Cluster()</weak_warning>;
        $id2 = <weak_warning descr="Operand types of '??' do not match ([\Node] vs [int]).">$nodes[$id] ?? $i</weak_warning>;
        $id3 = <weak_warning descr="Operand types of '??' do not match ([\Node] vs [int]).">$nodes[$id] ?? $counts[$id]</weak_warning>;
        return $owner->marginX;
    }
}
