<?php
namespace App;

use App\Mapping as ORM;

#[ORM\Entity('foo')]
final class Marker {}

#[Get(name: 'only')]
#[Post]
class Operation {}

class <weak_warning descr="This class declares no members; remove it or give it a purpose.">Plain</weak_warning> {}
