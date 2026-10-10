<?php
namespace Negative0 { match($x){1=>1,'1'=>2,default=>3}; }
namespace Negative1 { match($x){$a=>1,$b=>2}; }
namespace Negative2 { match($x){f()=>1,f()=>2}; }
namespace Negative3 { match($x){1=>1,2=>2}; }
