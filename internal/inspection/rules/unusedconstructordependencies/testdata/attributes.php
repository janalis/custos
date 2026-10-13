<?php
use Doctrine\ORM\Mapping as ORM;

class Bucket
{
    #[ORM\Id]
    #[ORM\Column(type: 'integer')]
    private int $id;

    private string $note;

    public function __construct(int $id, string $note)
    {
        $this->id = $id; // read by the ORM through reflection
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->note</weak_warning> = $note;
    }

    public function name(): string { return 'bucket'; }
}
