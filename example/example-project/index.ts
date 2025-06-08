import { GetSubtotal, goBootstrap } from "@polygo/example";

const products = [
  {
    id: 1,
    name: "Product 1",
    price: 100,
    category: {

      id: 1,
      name: "Electronics"
    }
  },
  {
    id: 2,
    name: "Product 2",
    price: 200,
    category: null
  }
];

async function main() {
  await goBootstrap();
  const result = GetSubtotal(products)
  console.log(result);
}

main();