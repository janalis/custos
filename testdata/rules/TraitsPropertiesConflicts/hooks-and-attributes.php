<?php
namespace Doctrine\ORM\Mapping {
    #[\Attribute]
    final class Column { public function __construct(mixed ...$args) {} }
}

namespace App {
    use Doctrine\ORM\Mapping as ORM;

    abstract class Node {
        public ?string $locale = null {
            set(?string $v) { $this->locale = $v === null ? null : strtolower($v); }
        }
        public ?string $code = null;
        public ?string $kind = null;
        public string $tag = '' {
            set(string $v) { $this->tag = strtolower($v); }
        }
    }
    trait Tagged {
        public string $tag = '';
    }
    trait Localised {
        #[ORM\Column(length: 35, nullable: true)]
        public ?string $locale = null;
    }
    trait Coded {
        public ?string $code = null;
    }
    trait Titled {
        public string $title = '';
    }
    trait Kinded {
        #[ORM\Column]
        public ?string $kind = 'x';
    }
    trait Hooked {
        public string $slug = '' {
            set(string $v) { $this->slug = strtolower($v); }
        }
        public int $rank = 0;
    }
    final class Page extends Node {
        use Localised, <weak_warning descr="Page and trait Coded both declare property $code.">Coded</weak_warning>, Titled, <error descr="Page and trait Kinded both declare property $kind.">Kinded</error>;

        public string <error descr="Page and trait Titled both declare property $title.">$title</error> = '' {
            set(string $v) { $this->title = trim($v); }
        }
    }
    final class Note extends Node {
        use <weak_warning descr="Note and trait Tagged both declare property $tag.">Tagged</weak_warning>; // parent hooks are ignored in check B
    }
    final class Post {
        use Hooked;

        public string <error descr="Post and trait Hooked both declare property $slug.">$slug</error> = '';

        public function __construct(public int <error descr="Post and trait Hooked both declare property $rank.">$rank</error> = 0 { get => $this->rank; }) {}
    }
}
