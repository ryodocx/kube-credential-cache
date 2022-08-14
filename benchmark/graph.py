import matplotlib.pyplot as plt


# kubectl
fig1, ax1 = plt.subplots()
x = ['local(no-wait)', 'local(cache)', 'local(wait)', 'eks(cache)', 'eks']
y = [153, 157, 650, 211, 812]
ax1.set_title("kubectl[ms]")
ax1.bar(x, y)
for x, y in zip(x, y):
    plt.text(x, y, y, ha='center', va='bottom')
fig1.savefig('graph_kubectl.svg')

# get credential
fig2, ax2 = plt.subplots()
x = ['local(no-wait)', 'local(cache)', 'local(wait)', 'eks(cache)', 'eks']
y = [0, 1.74, 496, 1.73, 540]
ax2.set_title("get credential[ms]")
ax2.bar(x, y)
for x, y in zip(x, y):
    plt.text(x, y, y, ha='center', va='bottom')
fig2.savefig('graph_credential.svg')

# eks kubectl
fig3, ax3 = plt.subplots()
x = ['with cache', 'normal']
y = [211, 812]
ax3.set_title("run kubectl with EKS[ms]")
ax3.bar(x, y)
for x, y in zip(x, y):
    plt.text(x, y, y, ha='center', va='bottom')
fig3.savefig('graph_eks.svg')
