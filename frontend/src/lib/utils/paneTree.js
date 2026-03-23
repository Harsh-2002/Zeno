export function forEachPane(node, cb) {
  if (!node) return;
  if (node.type === 'pane') { cb(node); return; }
  forEachPane(node.first, cb);
  forEachPane(node.second, cb);
}

export function findPaneById(node, id) {
  if (!node) return null;
  if (node.type === 'pane') return node.id === id ? node : null;
  return findPaneById(node.first, id) || findPaneById(node.second, id);
}

export function findParent(root, target) {
  if (!root || root.type === 'pane') return null;
  if (root.first === target) return { parent: root, which: 'first' };
  if (root.second === target) return { parent: root, which: 'second' };
  return findParent(root.first, target) || findParent(root.second, target);
}

export function countPanes(node) {
  if (!node) return 0;
  if (node.type === 'pane') return 1;
  return countPanes(node.first) + countPanes(node.second);
}

export function getSplitBadgeType(rootNode) {
  if (!rootNode || rootNode.type === 'pane') return null;
  const count = countPanes(rootNode);
  if (count <= 1) return null;
  if (count > 2) return 'multi';
  return rootNode.direction === 'vertical' ? 'v' : 'h';
}
