import rpyc
import time

class MyService(rpyc.Service):
    def on_connect(self, conn):
    # código que é executado quando uma conexão é iniciada, caso seja necessário
        pass

    def on_disconnect(self, conn):
    # código que é executado quando uma conexão é finalizada, caso seja necessário
        pass
    
    def exposed_get_answer(self): # este é um método exposto 
        return 42
    
    exposed_the_real_answer_though = 43 # este é um atributo exposto

    def exposed_get_question(self): # este método não é exposto
        return "Qual é a cor do cavalo branco de Napoleão?"
    
    def exposed_get_sum(self, vet): #Questão 3 e 4
        start = time.time()
        var_sum = 0
        for i in range(len(vet)):
            var_sum += vet[i]

        end = time.time()
        print(f'Tempo para executar o procedimento no servidor: {end-start}')

        return var_sum
    
#Para iniciar o servidor
if __name__ == "__main__":
    start = time.time()

    from rpyc.utils.server import ThreadedServer
    t = ThreadedServer(MyService, port=18861)
    t.start()

    end = time.time()
    print(end-start)