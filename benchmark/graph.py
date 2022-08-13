import matplotlib.pyplot as plt


fig1, ax1 = plt.subplots()
x = ['no-wait', 'cache', 'slow']
y = [158, 160, 701]
ax1.set_title("kubectl[ms]")
ax1.bar(x, y)
for x, y in zip(x, y):
    plt.text(x, y, y, ha='center', va='bottom')
fig1.savefig('graph_kubectl.svg')

fig2, ax2 = plt.subplots()
x = ['no-wait', 'cache', 'slow', 'eks']
y = [0, 1.74, 532, 531]
ax2.set_title("get credential[ms]")
ax2.bar(x, y)
for x, y in zip(x, y):
    plt.text(x, y, y, ha='center', va='bottom')
fig2.savefig('graph_credential.svg')

fig3, ax3 = plt.subplots()
x = ['cached', 'normal']
y = [220, 765]
ax3.set_title("run kubectl version(with EKS)[ms]")
ax3.bar(x, y)
for x, y in zip(x, y):
    plt.text(x, y, y, ha='center', va='bottom')
fig3.savefig('graph_eks.svg')
