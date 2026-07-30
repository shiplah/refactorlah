import { CheckoutCard } from '../features/checkout/old-card';
export { CheckoutCard as CardExport } from '../features/checkout/old-card';
const lazy = import('../features/checkout/old-card');
const common = require('../features/checkout/old-card');
import AliasCard from '@app/features/checkout/old-card';
import ViteCard from '@ui/features/checkout/old-card';
import PackageCard from '#app/features/checkout/old-card';
import SelfCard from '@fixture/shop/src/features/checkout/old-card';
import StableTypeScriptAlias from '@checkoutCard';
import StablePackageAlias from '#checkout-card';
import SimilarCard from '../features/checkout/old-card-extra';
import AmbiguousAlias from '@ambiguous';
import ConditionalPackageAlias from '#conditional/features/checkout/old-card';
const dynamic = import(`../features/checkout/${name}`);
const concatenated = require('../features/checkout/' + name);

export function render() {
  return CheckoutCard || AliasCard || ViteCard || PackageCard || SelfCard || StableTypeScriptAlias || StablePackageAlias || SimilarCard || AmbiguousAlias || ConditionalPackageAlias || lazy || common || dynamic || concatenated;
}
