<?php
namespace Doctrine\Common\Collections {
    interface Collection {}
    class ArrayCollection implements Collection {}
}

namespace App {
    use Doctrine\Common\Collections\ArrayCollection;
    use Doctrine\Common\Collections\Collection;

    class Role
    {
        /** @var ArrayCollection<int, string> */
        private $permissions;
        /** @var Collection<int, string> */
        private $members;

        public function __construct()
        {
            $this->permissions = new ArrayCollection();
            $this->members = new ArrayCollection();
        }

        /** @return ArrayCollection<int, string> */
        public function getPermissions()
        {
            return $this->permissions;
        }

        /** @return Collection<int, string> */
        public function <weak_warning descr="Declare ': Collection' as the return type.">getMembers</weak_warning>()
        {
            return $this->members;
        }
    }

    interface HasItems
    {
        /** @return ArrayCollection */
        public function items();
    }
}
